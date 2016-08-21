package experiment

type (
	// Result represents the result from running all the observations and comparing
	// them with the given compare method.
	Result struct {
		mismatches   []Observation
		candidates   []Observation
		observations Observations
		cm           ComparisonMethod
		hasRun       bool
	}
)

// NewResult takes an Observations type and will compare every test observation
// in it with the control observation through the given ComparisonMethod.
func NewResult(obs Observations, cm ComparisonMethod) *Result {
	return &Result{
		observations: obs,
		cm:           cm,
	}
}

// Mismatches returns all the observations for the tests that do not evaluate
// to true with the given ComparisonMethod.
// Note that this could potentially be an expensive method to run. It is advised
// to look at these results in a goroutine.
func (r *Result) Mismatches() []Observation {
	r.ensureRun()

	return r.mismatches
}

// Candidates returns all the observations for the tests that evaluate to true
// with the given ComparisonMethod.
// Note that this could potentially be an expensive method to run. It is advised
// to look at these results in a goroutine.
func (r *Result) Candidates() []Observation {
	r.ensureRun()

	return r.candidates
}

// Control returns the observation for the control test.
func (r *Result) Control() Observation {
	return r.observations.Control()
}

func (r *Result) ensureRun() {
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
