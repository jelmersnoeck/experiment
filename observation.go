package experiment

import "time"

type (
	Observation struct {
		Name     string
		Value    interface{}
		Error    error
		Panic    interface{}
		Duration time.Duration
	}

	// Observations resembles a set of observations
	Observations map[string]Observation
)

// Control returns the control observation from a set of observations.
func (o Observations) Control() Observation {
	return o.Find(controlKey)
}

// Candidates returns all the observations except the control one.
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
