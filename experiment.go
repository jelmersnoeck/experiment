package experiment

import "errors"

type (
	// Experiment is the experiment runner. It contains all the logic on how to run
	// experiments against controls and for a given number of users.
	Experiment struct {
		opts options
	}
)

var (
	NoNameError = errors.New("No name given for this experiment.")
)

// New will create a new Experiment and set it up for later usage. If a new
// experiment is created without name, an error will be returned.
func New(options ...Option) (*Experiment, error) {
	exp := &Experiment{opts: newOptions(options...)}

	if exp.Name() == "" {
		return nil, NoNameError
	}

	return exp, nil
}

func (e *Experiment) Name() string {
	return e.opts.name
}
