package experiment

import (
	"errors"
	"math/rand"
)

type (
	CandidateFunc func() (interface{}, error)
)

// Experiment represents a new refactoring experiment. This is where you'll
// define your control and candidates on and this will run the experiment
// according to the configuration.
type Experiment struct {
	Config *Config

	shouldRun    bool
	candidates   map[string]CandidateFunc
	observations map[string]Observation
}

// New creates a new Experiment with the given configuration options.
func New(cfgs ...ConfigFunc) *Experiment {
	cfg := &Config{}
	for _, c := range cfgs {
		c(cfg)
	}

	return &Experiment{
		Config:       cfg,
		shouldRun:    cfg.Percentage > 0 && rand.Intn(100) <= cfg.Percentage,
		candidates:   map[string]CandidateFunc{},
		observations: map[string]Observation{},
	}
}

// Before filter to do expensive setup only when the experiment is going to run.
// This will be skipped if the experiment doesn't need to run. A good use case
// would be to do a deep copy of a struct.
func (e *Experiment) Before(fnc func() error) {
}

// Control represents the control function, this resembles the old or current
// implementation. This function will always run, regardless of the
// configuration percentage or overwrites. If this function panics, the
// application will panic.
// The output of this function will be the base to what all the candidates will
// be compared to.
func (e *Experiment) Control(fnc CandidateFunc) {
	e.Candidate("control", fnc)
}

// Candidate represents a refactoring solution. The order of candidates is
// randomly determined.
// If the concurrent configuration is given, candidates will run concurrently.
// If a candidate panics, your application will not panic, but the candidate
// will be marked as failed.
func (e *Experiment) Candidate(name string, fnc CandidateFunc) {
	e.candidates[name] = fnc
}

// Compare represents the comparison functionality between a control and a
// candidate.
func (e *Experiment) Compare(fnc func(interface{}, interface{}) bool) {
}

// Clean will cleanup the state of a candidate (control included). This is done
// so the state could be cleaned up before storing for later comparison.
func (e *Experiment) Clean(fnc func(interface{})) {
}

// Force lets you overwrite the percentage. If set to true, the candidates will
// definitely run.
func (e *Experiment) Force(f bool) {
	if f == true {
		e.shouldRun = true
	}
}

// Ignore lets you decide if the experiment should be ignored this run or not.
// If set to true, the candidates will not run.
func (e *Experiment) Ignore(i bool) {
	if i == true {
		e.shouldRun = false
	}
}

// Run runs all the candidates and control in a random order. The value of the
// control function will be returned.
// If the concurrency configuration is given, this will return as soon as the
// control has finished running.
func (e *Experiment) Run() (interface{}, error) {
	// don't run the candidates, just the control
	if !e.shouldRun {
		fnc := e.candidates["control"]
		return fnc()
	}

	cChan := make(chan Observation)
	go e.run(cChan)

	select {
	case obs := <-cChan:
		return obs.Value, obs.Err
	}
}

func (e *Experiment) run(cChan chan Observation) {
	ack := e.ack()
	obsChan := make(chan Observation)

	for k, v := range e.candidates {
		go func(name string, fnc CandidateFunc) {
			ack <- true

			defer func() {
				if name == "control" {
					return
				}

				if r := recover(); r != nil {
					obsChan <- Observation{
						Name:  name,
						Value: r,
						Err:   errors.New("Panic"),
					}
				}
			}()

			v, err := fnc()

			obsChan <- Observation{
				Name:  name,
				Value: v,
				Err:   err,
			}
		}(k, v)
	}

	for range e.candidates {
		select {
		case obs := <-obsChan:
			if obs.Name == "control" {
				cChan <- obs
			}

			e.observations[obs.Name] = obs
		}
	}
}

func (e *Experiment) ack() chan bool {
	if e.Config.Concurrency {
		return make(chan bool, len(e.candidates))
	}

	return make(chan bool, 1)
}
