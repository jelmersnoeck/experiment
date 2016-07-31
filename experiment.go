package experiment

import (
	"errors"
	"fmt"
	"sync"

	"golang.org/x/net/context"
)

const (
	controlKey = "control"
)

type (
	// Experiment is the experiment runner. It contains all the logic on how to run
	// experiments against controls and for a given number of users.
	Experiment struct {
		sync.Mutex
		Config Config

		hits       float32
		runs       float32
		behaviours map[string]*behaviour
	}
)

var (
	// ErrMissingControl is returned when there is no control function set for the
	// experiment
	ErrMissingControl = errors.New("No control function was given.")
	// ErrNoControlObservation is returned in the case when there is something
	// wrong with running the control.
	ErrNoControlObservation = errors.New("The control did not finish properly.")
	// No Observations
	ErrNoObservations = errors.New("No observations could be generated")
	// ErrRunExperiment is returned when the experiment is required to have run
	// but has not run yet.
	ErrRunExperiment = errors.New("Experiment has not run yet, call `Run()` first.")
)

// New will create a new experiment with the given config.
func New(cfg Config) *Experiment {
	return &Experiment{Config: cfg}
}

// Control sets the control method for this experiment. The control should only
// be set once and this will return an error if this is not the case.
func (e *Experiment) Control(b BehaviourFunc) error {
	return e.Test(controlKey, b)
}

// Test adds a test case to the exeriment. If a test case with the same name is
// already used, an error will be returned.
func (e *Experiment) Test(name string, b BehaviourFunc) error {
	e.Lock()
	defer e.Unlock()

	if e.behaviours == nil {
		e.behaviours = map[string]*behaviour{}
	}

	if _, ok := e.behaviours[name]; ok {
		return fmt.Errorf("Behaviour `%s` already exists.", name)
	}

	e.behaviours[name] = newBehaviour(name, b)
	return nil
}

// Run will go through all the tests in a random order and run them one by one.
// The Result that is returned contains the Control and Candidate behaviours.
func (e *Experiment) Run(ctx context.Context) (Observations, error) {
	if _, ok := e.behaviours[controlKey]; !ok {
		return Observations{}, ErrMissingControl
	}

	defer func() {
		e.Lock()
		e.runs++
		e.Unlock()
	}()

	runner := &experimentRunner{}

	if e.shouldRun() {
		e.Lock()
		e.hits++
		e.Unlock()

		return runner.run(ctx, e.Config.BeforeFilters, e.behaviours), nil
	}

	beh := map[string]*behaviour{
		controlKey: e.behaviours[controlKey],
	}
	return runner.run(ctx, e.Config.BeforeFilters, beh), nil
}

func (e *Experiment) shouldRun() bool {
	e.Lock()
	defer e.Unlock()

	if TestMode {
		return true
	}

	if hitRate := (e.hits / e.runs) * 100.0; hitRate > e.Config.Percentage {
		return false
	}

	return true
}
