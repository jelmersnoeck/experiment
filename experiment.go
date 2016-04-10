package experiment

import (
	"errors"
	"fmt"
	"sync"
)

type (
	// Experiment is the experiment runner. It contains all the logic on how to run
	// experiments against controls and for a given number of users.
	Experiment struct {
		*sync.Mutex
		opts       options
		behaviours map[string]*behaviour
	}
)

var (
	NoNameError = errors.New("No name given for this experiment.")
)

// New will create a new Experiment and set it up for later usage. If a new
// experiment is created without name, an error will be returned.
func New(options ...Option) (*Experiment, error) {
	exp := &Experiment{
		Mutex:      &sync.Mutex{},
		opts:       newOptions(options...),
		behaviours: map[string]*behaviour{},
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

func (e *Experiment) Name() string {
	return e.opts.name
}
