package experiment

import "time"

type (
	// Observation is the result of a test being run.
	Observation interface {
		// Value is the value returned by the test or control function that gets
		// executed.
		Value() interface{}
		// Error is the error that gets returned when executing the test or
		// control function.
		Error() error
		// Panic is the panic that occured whilst running a test.
		Panic() interface{}
		Name() string
		Duration() time.Duration
	}

	experimentObservation struct {
		name     string
		value    interface{}
		err      error
		panic    interface{}
		duration time.Duration
	}
)

func (o *experimentObservation) Name() string {
	return o.name
}

func (o *experimentObservation) Value() interface{} {
	return o.value
}

func (o *experimentObservation) Error() error {
	return o.err
}

func (o *experimentObservation) Panic() interface{} {
	return o.panic
}

func (o *experimentObservation) Duration() time.Duration {
	return o.duration
}
