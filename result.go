package experiment

type (
	// Result represents the result from running all the observations and comparing
	// them with the given compare method.
	Result interface {
		// Mismatches represent the observations on which the given `Compare`
		// option returned false.
		Mismatches() []Observation
		// Candidates represents all the observations which are not the control.
		Candidates() []Observation
		// Control represents the observation of the results from the control
		// function.
		Control() Observation
	}

	experimentResult struct {
		mismatches []Observation
		candidates []Observation
		control    Observation
		experiment *Experiment
	}
)

// NewResult will take an experiment and go over all the observations to find
// if the observation is a match or a mismatch.
func NewResult(e *Experiment) Result {
	rs := &experimentResult{
		mismatches: []Observation{},
		candidates: []Observation{},
		experiment: e,
	}

	for _, o := range e.observations {
		if o.Name() == "control" {
			rs.control = o
		} else {
			rs.candidates = append(rs.candidates, o)
		}
	}

	if rs.experiment.opts.comparison != nil {
		rs.evaluate()
	}

	return rs
}

func (r *experimentResult) Mismatches() []Observation {
	return r.mismatches
}

func (r *experimentResult) Candidates() []Observation {
	return r.candidates
}

func (r *experimentResult) Control() Observation {
	return r.control
}

// evaluate does the hard work of going through all the result Candidates and
// calls the compare method given to the experiment with the Control and the
// candidate value. If no comparison method is given in the options, evaluate
// is skipped.
func (r *experimentResult) evaluate() {
	for _, c := range r.candidates {
		if !r.experiment.opts.comparison(r.control, c) {
			r.mismatches = append(r.mismatches, c)
		}
	}
}
