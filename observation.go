package experiment

import "time"

type (
	// Observation is the result of running a test. It contains the information
	// that is obtained by calling one of the test functions.
	Observation struct {
		// Name is the name of the associated test for this observation.
		Name string
		// Value is the outcome of the test that has run. This is a result of a
		// BebaviourFunc.
		Value interface{}
		// Error is an error that might have occurred while running the test.
		// This is a result of a BehaviourFunc.
		Error error
		// Panic is a panic that might have occurred whilst running the test.
		// Panics are only associated with tests. The Control function will
		// panic as it would be doing without the experiment wrapper.
		Panic interface{}
		// Duration is the time it takes to run the test behaviour.
		Duration time.Duration
	}

	// Observations resembles a set of observations with some extra access
	// functionality.
	Observations map[string]Observation
)

// Control returns the control observation from a set of observations.
func (o Observations) Control() Observation {
	return o.Find(controlKey)
}

// Tests returns all the observations except the control one.
func (o Observations) Tests() []Observation {
	var os []Observation
	for key, obs := range o {
		if key == controlKey {
			continue
		}

		os = append(os, obs)
	}
	return os
}

// Find returns an observation for the test with the given name. If there is no
// such test, nil will be returned.
func (o Observations) Find(name string) Observation {
	return o[name]
}
