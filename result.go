package experiment

// Result represents the result from running all the observations and comparing
// them with the given compare method.
type experimentResult struct {
	Mismatches []Observation
	Candidates []Observation
	Control    Observation

	experiment *Experiment
}

// NewResult will take an experiment and go over all the observations to find
// if the observation is a match or a mismatch.
func NewResult(e *Experiment) *experimentResult {
	rs := &experimentResult{
		Mismatches: []Observation{},
		Candidates: []Observation{},
		experiment: e,
	}

	for _, o := range e.observations {
		if o.Name() == "control" {
			rs.Control = o
		} else {
			rs.Candidates = append(rs.Candidates, o)
		}
	}

	if rs.experiment.opts.comparison != nil {
		rs.evaluate()
	}

	return rs
}

// evaluate does the hard work of going through all the result Candidates and
// calls the compare method given to the experiment with the Control and the
// candidate value. If no comparison method is given in the options, evaluate
// is skipped.
func (r *experimentResult) evaluate() {
	for _, c := range r.Candidates {
		if !r.experiment.opts.comparison(r.Control, c) {
			r.Mismatches = append(r.Mismatches, c)
		}
	}
}
