package experiment

import "fmt"

// ErrCandidatePanic represents the error that a candidate panicked.
type CandidatePanicError struct {
	Name  string
	Panic interface{}
}

// Error returns a simple error message. It does not include the panic information.
func (e CandidatePanicError) Error() string {
	return fmt.Sprintf("experiment candidate '%s' panicked", e.Name)
}
