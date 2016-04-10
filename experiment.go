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
