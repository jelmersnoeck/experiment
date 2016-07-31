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

// NewResult will take an experiment and go over all the observations to find
// if the observation is a match or a mismatch.
func NewResult(obs Observations, cm ComparisonMethod) Result {
	return &experimentResult{
		observations: obs,
		cm:           cm,
	}
}

func (r *experimentResult) Mismatches() []Observation {
	return r.mismatches
}

func (r *experimentResult) Candidates() []Observation {
	return r.observations.Candidates()
}

func (r *experimentResult) Control() Observation {
	return r.observations.Control()
}

func (r *experimentResult) Publish() error {
	return nil
}
