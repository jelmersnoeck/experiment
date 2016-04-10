package experiment

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/net/context"
)

type (
	// Experiment is the experiment runner. It contains all the logic on how to run
	// experiments against controls and for a given number of users.
	Experiment struct {
		*sync.Mutex
		opts         options
		behaviours   map[string]*behaviour
		observations map[string]Observation
		rand         *rand.Rand
		runs         float64
		hits         float64
		ctx          context.Context
	}
)

var (
	NoNameError          = errors.New("No name given for this experiment.")
	MissingControlError  = errors.New("No control function was given.")
	MissingTestError     = errors.New("No test function was given.")
	NoControlObservation = errors.New("The control did not finish properly.")
	RunExperimentError   = errors.New("Experiment has not run yet, call `Run()` first.")

	defaultOptions = []Option{}
)

// Init is used to set default options. If you have several experiments running
// and would like to set some default options, this is the way to go. Any
// option given to the `New()` function will overwrite the default option.
//
// This can also be used to mark the setup for testing.
// TODO: make separation between Init options and New options.
func Init(options ...Option) {
	defaultOptions = append(defaultOptions, options...)
}

// New will create a new Experiment and set it up for later usage.
func New(nm string, options ...Option) *Experiment {
	ops := defaultOptions
	ops = append(ops, options...)
	ops = append(ops, name(nm))
	opts := newOptions(ops...)
	exp := &Experiment{
		Mutex:        &sync.Mutex{},
		opts:         opts,
		behaviours:   map[string]*behaviour{},
		observations: map[string]Observation{},
		rand:         rand.New(rand.NewSource(time.Now().UnixNano())),
		ctx:          opts.ctx,
	}

	return exp
}

func (e *Experiment) Name() string {
	return e.opts.name
}

// Control sets the control method for this experiment. The control should only
// be set once and this will return an error if this is not the case.
func (e *Experiment) Control(b BehaviourFunc) error {
	return e.Test("control", b)
}

// Test adds a test case to the exeriment. If a test case with the same name is
// already used, an error will be returned.
func (e *Experiment) Test(name string, b BehaviourFunc) error {
	if _, ok := e.behaviours[name]; ok {
		return errors.New(fmt.Sprintf("Behaviour `%s` already exists.", name))
	}

	e.Lock()
	e.behaviours[name] = newBehaviour(name, b)
	e.Unlock()

	return nil
}

// Result returns a Result type created from the observations made running the
// experiment. This method should be called within a goroutine as it is an
// expensive method to execute. If the test has not run yet, an error will be
// returned. The `Run` method is expected to be used within the application
// and thus should not be part of `Result`.
func (e *Experiment) Result() (*experimentResult, error) {
	if len(e.observations) == 0 {
		return nil, RunExperimentError
	}

	return NewResult(e), nil
}

// Run will go through all the tests in a random order and run them one by one.
// After all the tests have run, it will use the Observation for the control
// behaviour.
func (e *Experiment) Run() (Observation, error) {
	defer func() {
		e.Lock()
		e.runs += 1
		e.Unlock()
	}()

	if _, ok := e.behaviours["control"]; !ok {
		return nil, MissingControlError
	}

	if len(e.behaviours) < 2 {
		return nil, MissingTestError
	}

	if !e.shouldRun() {
		for _, b := range e.behaviours {
			if b.name == "control" {
				obs := &experimentObservation{name: "control"}
				return e.makeObservation(b, obs), nil
			}
		}

		return nil, NoControlObservation
	}

	// if we reach this point, it means we should hit the tests
	defer func() {
		e.Lock()
		e.hits += 1
		e.Unlock()
	}()

	for _, bef := range e.opts.before {
		e.ctx = bef(e.ctx)
	}

	bhs := []*behaviour{}
	for _, b := range e.behaviours {
		bhs = append(bhs, b)
	}
	e.observe(bhs)

	for _, o := range e.observations {
		if o.Name() == "control" {
			return o, nil
		}
	}

	return nil, NoControlObservation
}

func (e *Experiment) shouldRun() bool {
	if e.opts.testMode {
		return true
	}

	if !e.opts.enabled {
		return false
	}

	pct := e.runs / e.hits
	if pct > e.opts.percentage {
		return false
	}

	return true
}

// observe is the actual runner that goes through a list of behaviours and
// executes them. It will do so in a random order.
//
// For safety purpose, all functions that are not the control are run in a
// goroutine with a recover function. This way, when a panic would occur in one
// of the tests, the user would not notice. However, if a panic happens in the
// control, it will actually be triggered. This happens after we collect all
// the data.
func (e *Experiment) observe(behaviours []*behaviour) {
	for _, key := range e.rand.Perm(len(behaviours)) {
		var wg sync.WaitGroup
		wg.Add(1)
		go func(wg *sync.WaitGroup, b *behaviour) {
			obs := &experimentObservation{name: b.name}
			defer func() {
				wg.Done()
				// If the control throws a panic, the application should deal
				// with this panic. The tests should never have an impact on the
				// user, so for all the other behaviours we'll add a recover.
				// The second case is when we're in test mode. Within a test,
				// we always want to know if something gives us a panic or not.
				if obs.Name() == "control" || e.opts.testMode {
					return
				} else if r := recover(); r != nil {
					obs.panic = r
				}
			}()

			e.makeObservation(b, obs)
		}(&wg, behaviours[key])
		wg.Wait()
	}
}

func (e *Experiment) makeObservation(b *behaviour, obs *experimentObservation) Observation {
	start := time.Now()
	defer func() {
		obs.duration = time.Now().Sub(start)
	}()

	e.Lock()
	e.observations[b.name] = obs
	e.Unlock()

	obs.value, obs.err = b.fnc(e.ctx)
	return obs
}
