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
		mismatches   []Observation
		candidates   []Observation
		observations Observations
		cm           ComparisonMethod
		hasRun       bool
	}
)

// NewResult takes an Observations type and will compare every test observation
// in it with the control observation through the given ComparisonMethod.
func NewResult(obs Observations, cm ComparisonMethod) Result {
	return &experimentResult{
		observations: obs,
		cm:           cm,
	}
}

func (r *experimentResult) Mismatches() []Observation {
	r.ensureRun()

	return r.mismatches
}

func (r *experimentResult) Candidates() []Observation {
	r.ensureRun()

	return r.candidates
}

func (r *experimentResult) Control() Observation {
	return r.observations.Control()
}

func (r *experimentResult) ensureRun() {
	if r.hasRun {
		return
	}
	defer func() { r.hasRun = true }()

	if r.cm == nil {
		r.candidates = r.observations.Tests()
		return
	}

	ctrl := r.observations.Control()
	for _, obs := range r.observations.Tests() {
		if r.cm(ctrl, obs) {
			r.candidates = append(r.candidates, obs)
		} else {
			r.mismatches = append(r.mismatches, obs)
		}
	}

}
