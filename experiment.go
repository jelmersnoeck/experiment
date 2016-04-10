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
		observations map[string]*Observation
		rand         *rand.Rand
	}
)

var (
	NoNameError         = errors.New("No name given for this experiment.")
	MissingControlError = errors.New("No control function was given.")
	MissingTestError    = errors.New("No test function was given.")
)

// New will create a new Experiment and set it up for later usage. If a new
// experiment is created without name, an error will be returned.
func New(options ...Option) (*Experiment, error) {
	exp := &Experiment{
		Mutex:        &sync.Mutex{},
		opts:         newOptions(options...),
		behaviours:   map[string]*behaviour{},
		observations: map[string]*Observation{},
		rand:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	if exp.Name() == "" {
		return nil, NoNameError
	}

	return exp, nil
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

// Run will go through all the tests in a random order and run them one by one.
// After all the tests have run, it will use the Observation for the control
// behaviour.
func (e *Experiment) Run() (*Observation, error) {
	if _, ok := e.behaviours["control"]; !ok {
		return nil, MissingControlError
	}

	if len(e.behaviours) < 2 {
		return nil, MissingTestError
	}

	bhs := []*behaviour{}
	for _, b := range e.behaviours {
		bhs = append(bhs, b)
	}
	e.observe(bhs)

	return e.observations["control"], nil
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
		go func(wg *sync.WaitGroup, b *behaviour, e *Experiment) {
			start := time.Now()
			obs := &Observation{Name: b.name}
			defer func() {
				wg.Done()
				obs.Duration = time.Now().Sub(start)

				// If the control throws a panic, the application should deal
				// with this panic. The tests should never have an impact on the
				// user, so for all the other behaviours we'll add a recover.
				if obs.Name == "control" {
					return
				} else if r := recover(); r != nil {
					obs.Panic = r
				}
			}()

			e.Lock()
			e.observations[b.name] = obs
			e.Unlock()

			obs.Value, obs.Error = b.fnc(context.Background())
		}(&wg, behaviours[key], e)
		wg.Wait()
	}
}

func (e *Experiment) Name() string {
	return e.opts.name
}
