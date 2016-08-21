package experiment

import (
	"time"

	"golang.org/x/net/context"
)

// Runner represents the implementation that actually runs the tests. Runners
// are not safe for concurrent use. Each concurrent request should request
// a new runner from the experiment.
type Runner struct {
	experiment *Experiment
	behaviours map[string]*behaviour

	config Config

	testMode bool
	disabled bool
	force    bool
	hasRun   bool

	hits float32
	runs float32
}

// Disable forces the runner to not run the tests. This overrules the `Force()`
// method.
func (r *Runner) Disable(d bool) {
	r.disabled = d
}

// Force forces the runner to run the tests no matter what the hitrate is or
// what other options are given.
func (r *Runner) Force(f bool) {
	r.force = f
}

// HasRun returns wether or not the test has actually been executed.
func (r *Runner) HasRun() bool {
	return r.hasRun
}

// Run will run the tests with a given context.
func (r *Runner) Run(ctx context.Context) Observations {
	if ctx == nil {
		ctx = context.Background()
	}

	for _, f := range r.config.BeforeFilters {
		ctx = f(ctx)
	}

	var behaviours map[string]*behaviour

	if r.disabled {
		// only run the control, we don't want to run the tests
		behaviours = map[string]*behaviour{
			controlKey: r.behaviours[controlKey],
		}
	} else if r.force {
		// we don't want to count a force run towards our percentage
		behaviours = r.behaviours
	} else {
		r.experiment.run()
		if r.shouldRun() {
			r.experiment.hit()
			r.hasRun = true
			behaviours = r.behaviours
		} else {
			behaviours = map[string]*behaviour{
				controlKey: r.behaviours[controlKey],
			}
		}
	}

	obsch := make(chan *Observation, len(behaviours))
	for _, beh := range behaviours {
		go r.observe(ctx, beh, obsch, TestMode)
	}

	obs := Observations{}
	for range behaviours {
		select {
		case ob := <-obsch:
			obs[ob.Name] = *ob
		}
	}

	return obs
}

func (r *Runner) shouldRun() bool {
	if r.testMode {
		return true
	}

	if r.runs == 0 {
		return true
	}

	if hitRate := (r.hits / r.runs) * 100; hitRate <= r.config.Percentage {
		return true
	}

	return false
}

func (r *Runner) observe(ctx context.Context, beh *behaviour, obsch chan *Observation, tm bool) {
	obs := &Observation{Name: beh.name}

	defer func() {
		// If the control throws a panic, the application should deal
		// with this panic. The tests should never have an impact on the
		// user, so for all the other behaviours we'll add a recover.
		// When we're in TestMode, we shouldn't skip panics either.
		if !(obs.Name == controlKey || tm) {
			if r := recover(); r != nil {
				obs.Panic = r
			}
		}

		obsch <- obs
	}()

	runObservation(ctx, beh, obs)
}

func runObservation(ctx context.Context, b *behaviour, obs *Observation) {
	defer func(start time.Time) {
		obs.Duration = time.Now().Sub(start)
	}(time.Now())
	obs.Value, obs.Error = b.fnc(ctx)
}
