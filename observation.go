package experiment

import "time"

// Observation represents the outcome of a candidate that has run.
type Observation[C any] struct {
	Duration     time.Duration
	Error        error
	Success      bool
	Name         string
	Panic        interface{}
	Value        C
	CleanValue   C
	ControlValue C
}
