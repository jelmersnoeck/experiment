package experiment

import (
	"sync"
	"time"

	"golang.org/x/net/context"
)

type experimentRunner struct {
	sync.Mutex
}

// observe is the actual runner that goes through a list of behaviours and
// executes them. It will do so in a random order.
//
// For safety purpose, all functions that are not the control are run in a
// goroutine with a recover function. This way, when a panic would occur in one
// of the tests, the user would not notice. However, if a panic happens in the
// control, it will actually be triggered. This happens after we collect all
// the data.
func (r *experimentRunner) run(ctx context.Context, filters []BeforeFilter, behaviours map[string]*behaviour) Observations {
	for _, f := range filters {
		ctx = f(ctx)
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

func (r *experimentRunner) observe(ctx context.Context, beh *behaviour, obsch chan *Observation, tm bool) {
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
