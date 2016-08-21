package experiment

import (
	"errors"
	"fmt"
	"sync"
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
)

// New will create a new experiment with the given config. Experiments are safe
// for concurrent usage.
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

// Runner will return a new Runenr instance which can be used to run the actual
// experiment. A runner is not safe for concurrent usage, but this method is.
// At the point of calling this method, we will copy the hitrate, which means
// that any actual experiment runs (from other runners) that happen between
// requesting a new runner and actually running the test will not influence it's
// state.
func (e *Experiment) Runner() (*Runner, error) {
	if _, ok := e.behaviours[controlKey]; !ok {
		return nil, ErrMissingControl
	}

	return &Runner{
		experiment: e,
		config:     e.Config,
		behaviours: e.behaviours,
		testMode:   TestMode,
		hits:       e.hits,
		runs:       e.runs,
	}, nil
}

func (e *Experiment) hit() {
	e.Lock()
	e.hits++
	e.Unlock()
}

func (e *Experiment) run() {
	e.Lock()
	e.runs++
	e.Unlock()
}
