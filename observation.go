package experiment

import "time"

// Observation is the result of a test being run.
type Observation struct {
	Name     string
	Value    interface{}
	Error    error
	Panic    interface{}
	Duration time.Duration
}
