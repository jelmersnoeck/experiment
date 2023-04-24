package experiment

import (
	"math/rand"
	"time"
)

type (
	// BeforeFunc represents the function that gets run before the experiment
	// starts. This function will only run if the experiment should run. The
	// functionality should be defined by the user.
	BeforeFunc func() error

	// CandidateFunc represents a function that is implemented by a candidate.
	// The value returned is the value that will be used to compare data.
	CandidateFunc[C any] func() (C, error)

	// CleanFunc represents the function that cleans up the output data. This
	// function will only be called for candidates that did not error.
	CleanFunc[C any] func(C) C

	// CompareFunc represents the function that takes two candidates and knows
	// how to compare them. The functionality is implemented by the user. This
	// function will only be called for candidates that did not error.
	CompareFunc[C any] func(C, C) bool
)

// Experiment represents a new refactoring experiment. This is where you'll
// define your control and candidates on and this will run the experiment
// according to the configuration.
type Experiment[C any] struct {
	config    *Config
	publisher Publisher[C]

	shouldRun    bool
	candidates   map[string]CandidateFunc[C]
	observations map[string]*Observation[C]

	before  BeforeFunc
	compare CompareFunc[C]
	clean   CleanFunc[C]
}

// New creates a new Experiment with the given configuration options.
func New[C any](cfgs ...ConfigFunc) *Experiment[C] {
	cfg := &Config{}
	for _, c := range cfgs {
		c(cfg)
	}

	return &Experiment[C]{
		config:       cfg,
		shouldRun:    cfg.Percentage > 0 && rand.Intn(100) <= cfg.Percentage,
		candidates:   map[string]CandidateFunc[C]{},
		observations: map[string]*Observation[C]{},
	}
}

// WithPublisher configures the publisher for the experiment. The publisher must
// have the same type associated as the experiment.
func (e *Experiment[C]) WithPublisher(pub Publisher[C]) *Experiment[C] {
	e.publisher = pub
	return e
}

// Before filter to do expensive setup only when the experiment is going to run.
// This will be skipped if the experiment doesn't need to run. A good use case
// would be to do a deep copy of a struct.
func (e *Experiment[C]) Before(fnc BeforeFunc) {
	e.before = fnc
}

// Control represents the control function, this resembles the old or current
// implementation. This function will always run, regardless of the
// configuration percentage or overwrites. If this function panics, the
// application will panic.
// The output of this function will be the base to what all the candidates will
// be compared to.
func (e *Experiment[C]) Control(fnc CandidateFunc[C]) {
	e.candidates["control"] = fnc
}

// Candidate represents a refactoring solution. The order of candidates is
// randomly determined.
// If the concurrent configuration is given, candidates will run concurrently.
// If a candidate panics, your application will not panic, but the candidate
// will be marked as failed.
// If the name is control, this will panic.
func (e *Experiment[C]) Candidate(name string, fnc CandidateFunc[C]) error {
	if name == "control" {
		panic("can't use a candidate with the name 'control'")
	}

	e.candidates[name] = fnc
	return nil
}

// Compare represents the comparison functionality between a control and a
// candidate.
func (e *Experiment[C]) Compare(fnc CompareFunc[C]) {
	e.compare = fnc
}

// Clean will cleanup the state of a candidate (control included). This is done
// so the state could be cleaned up before storing for later comparison.
func (e *Experiment[C]) Clean(fnc CleanFunc[C]) {
	e.clean = fnc
}

// Force lets you overwrite the percentage. If set to true, the candidates will
// definitely run.
func (e *Experiment[C]) Force(f bool) {
	if f {
		e.shouldRun = true
	}
}

// Ignore lets you decide if the experiment should be ignored this run or not.
// If set to true, the candidates will not run.
func (e *Experiment[C]) Ignore(i bool) {
	if i {
		e.shouldRun = false
	}
}

// Run runs all the candidates and control in a random order. The value of the
// control function will be returned.
// If the concurrency configuration is given, this will return as soon as the
// control has finished running.
func (e *Experiment[C]) Run() (C, error) {
	// don't run the candidates, just the control
	if !e.shouldRun {
		fnc := e.candidates["control"]
		return fnc()
	}

	if e.before != nil {
		if err := e.before(); err != nil {
			var r C
			return r, err
		}
	}

	cChan := make(chan *Observation[C])
	go e.run(cChan)

	obs := <-cChan
	return obs.Value, obs.Error
}

func (e *Experiment[C]) run(cChan chan *Observation[C]) {
	if e.config.Concurrency {
		e.runConcurrent(cChan)
	} else {
		e.runSequential(cChan)
	}
}

func (e *Experiment[C]) runConcurrent(cChan chan *Observation[C]) {
	obsChan := make(chan *Observation[C])
	for k, v := range e.candidates {
		go func(name string, fnc CandidateFunc[C]) {
			runCandidate(name, fnc, obsChan)
		}(k, v)
	}

	for range e.candidates {
		obs := <-obsChan

		if obs.Name == "control" {
			cChan <- obs
		}
		e.observations[obs.Name] = obs
	}

	e.conclude()
}

func (e *Experiment[C]) runSequential(cChan chan *Observation[C]) {
	obsChan := make(chan *Observation[C])
	for k, v := range e.candidates {
		go func(name string, fnc CandidateFunc[C]) {
			runCandidate(k, v, obsChan)
		}(k, v)

		obs := <-obsChan
		e.observations[obs.Name] = obs
	}

	e.conclude()

	cChan <- e.observations["control"]
}

func (e *Experiment[C]) conclude() {
	control := e.observations["control"]

	for _, o := range e.observations {
		if o.Error == nil {
			if e.clean != nil {
				o.CleanValue = e.clean(o.Value)
			} else {
				o.CleanValue = o.Value
			}
		}
	}

	if e.compare != nil {
		for k, o := range e.observations {
			if o.Error == nil {
				if k == "control" {
					o.Success = true
					continue
				}

				o.Success = e.compare(control.Value, o.Value)
				o.ControlValue = control.CleanValue
			}
		}
	}

	if e.publisher != nil {
		for _, o := range e.observations {
			e.publisher.Publish(*o)
		}
	}
}

func runCandidate[C any](name string, fnc CandidateFunc[C], obsChan chan *Observation[C]) {
	start := time.Now()

	defer func() {
		if name == "control" {
			return
		}

		if r := recover(); r != nil {
			end := time.Now()

			obsChan <- &Observation[C]{
				Name: name,
				Error: CandidatePanicError{
					Name:  name,
					Panic: r,
				},
				Duration: end.Sub(start),
			}
		}
	}()

	v, err := fnc()
	end := time.Now()

	obsChan <- &Observation[C]{
		Name:     name,
		Value:    v,
		Error:    err,
		Duration: end.Sub(start),
	}
}
