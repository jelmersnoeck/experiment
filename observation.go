package experiment

// Observation represents the outcome of a candidate that has run.
type Observation struct {
	Name  string
	Value interface{}
	Err   error
}
