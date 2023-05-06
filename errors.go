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

// PublishError is an error used when the publisher returns an error. It
// combines all errors into a single error.
type PublishError struct {
	errors []error
}

func (e *PublishError) Error() string {
	var b []byte
	for i, err := range e.errors {
		if i > 0 {
			b = append(b, '\n')
		}
		b = append(b, err.Error()...)
	}
	return string(b)
}

func (e *PublishError) Unwrap() []error {
	return e.errors
}

func (e *PublishError) append(err error) {
	if err != nil {
		e.errors = append(e.errors, err)
	}
}
