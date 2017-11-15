package experiment

import (
	"errors"
	"math/rand"
	"time"
)

type (
	// BeforeFunc represents the function that gets run before the experiment
	// starts. This function will only run if the experiment should run. The
	// functionality should be defined by the user.
	BeforeFunc func() error

	// CaniddateFunc represents a function that is implemented by a candidate.
	// The value returned is the value that will be used to compare data.
	CandidateFunc func() (interface{}, error)

	// CleanFunc represents the function that cleans up the output data. This
	// function will only be called for candidates that did not error.
	CleanFunc func(interface{}) interface{}

	// CompareFunc represents the function that takes two candidates and knows
	// how to compare them. The functionality is implemented by the user. This
	// function will only be called for candidates that did not error.
	CompareFunc func(interface{}, interface{}) bool
)

var (
	// ErrControlCandidate is returned when a candidate is initiated with
	// control as it's name.
	ErrControlCandidate = errors.New("Can't use a candidate with the name 'control'")

	// ErrCandidatePanic represents the error that a candidate panicked.
	ErrCandidatePanic = errors.New("Candidate panicked.")
)

// Experiment represents a new refactoring experiment. This is where you'll
// define your control and candidates on and this will run the experiment
// according to the configuration.
type Experiment struct {
	Config *Config

	shouldRun    bool
	candidates   map[string]CandidateFunc
	observations map[string]*Observation

	before  BeforeFunc
	compare CompareFunc
	clean   CleanFunc
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
		observations: map[string]*Observation{},
	}
}

// Before filter to do expensive setup only when the experiment is going to run.
// This will be skipped if the experiment doesn't need to run. A good use case
// would be to do a deep copy of a struct.
func (e *Experiment) Before(fnc BeforeFunc) {
	e.before = fnc
}

// Control represents the control function, this resembles the old or current
// implementation. This function will always run, regardless of the
// configuration percentage or overwrites. If this function panics, the
// application will panic.
// The output of this function will be the base to what all the candidates will
// be compared to.
func (e *Experiment) Control(fnc CandidateFunc) {
	e.candidates["control"] = fnc
}

// Candidate represents a refactoring solution. The order of candidates is
// randomly determined.
// If the concurrent configuration is given, candidates will run concurrently.
// If a candidate panics, your application will not panic, but the candidate
// will be marked as failed.
// If the name is control, this will error and not add the candidate.
func (e *Experiment) Candidate(name string, fnc CandidateFunc) error {
	if name == "control" {
		return ErrControlCandidate
	}

	e.candidates[name] = fnc
	return nil
}

// Compare represents the comparison functionality between a control and a
// candidate.
func (e *Experiment) Compare(fnc CompareFunc) {
	e.compare = fnc
}

// Clean will cleanup the state of a candidate (control included). This is done
// so the state could be cleaned up before storing for later comparison.
func (e *Experiment) Clean(fnc CleanFunc) {
	e.clean = fnc
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

	if e.before != nil {
		e.before()
	}

	cChan := make(chan *Observation)
	go e.run(cChan)

	select {
	case obs := <-cChan:
		return obs.Value, obs.Error
	}
}

func (e *Experiment) run(cChan chan *Observation) {
	if e.Config.Concurrency {
		e.runConcurrent(cChan)
	} else {
		e.runSequential(cChan)
	}
}

func (e *Experiment) runConcurrent(cChan chan *Observation) {
	obsChan := make(chan *Observation)
	for k, v := range e.candidates {
		go func(name string, fnc CandidateFunc) {
			runCandidate(name, fnc, obsChan)
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

	e.conclude()
}

func (e *Experiment) runSequential(cChan chan *Observation) {
	obsChan := make(chan *Observation)
	for k, v := range e.candidates {
		go func(name string, fnc CandidateFunc) {
			runCandidate(k, v, obsChan)
		}(k, v)

		select {
		case obs := <-obsChan:
			e.observations[obs.Name] = obs
		}
	}

	e.conclude()

	cChan <- e.observations["control"]
}

func (e *Experiment) conclude() {
	control := e.observations["control"]

	if e.compare != nil {
		for k, o := range e.observations {
			if k == "control" {
				continue
			}

			if o.Error == nil {
				o.Success = e.compare(control.Value, o.Value)
			}
		}
	}

	for _, o := range e.observations {
		if o.Error == nil {
			if e.clean != nil {
				o.CleanValue = e.clean(o.Value)
			} else {
				o.CleanValue = o.Value
			}
		}
	}

	if e.Config.Publisher != nil {
		for _, o := range e.observations {
			e.Config.Publisher.Publish(*o)
		}
	}
}

func runCandidate(name string, fnc CandidateFunc, obsChan chan *Observation) {
	start := time.Now()

	defer func() {
		if name == "control" {
			return
		}

		if r := recover(); r != nil {
			end := time.Now()

			obsChan <- &Observation{
				Name:     name,
				Panic:    r,
				Error:    ErrCandidatePanic,
				Duration: end.Sub(start),
			}
		}
	}()

	v, err := fnc()
	end := time.Now()

	obsChan <- &Observation{
		Name:     name,
		Value:    v,
		Error:    err,
		Duration: end.Sub(start),
	}
}
