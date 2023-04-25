package experiment_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jelmersnoeck/experiment"
)

func TestRun(t *testing.T) {
	t.Run("sequential", func(t *testing.T) {
		t.Run("basic", func(t *testing.T) {
			testRun(t)
		})

		t.Run("timeouts", func(t *testing.T) {
			testWithTimeout(t)
		})
	})

	t.Run("concurrent", func(t *testing.T) {
		t.Run("basic", func(t *testing.T) {
			testRun(t, experiment.WithConcurrency())
		})

		t.Run("timeouts", func(t *testing.T) {
			testWithTimeout(t, experiment.WithConcurrency())
		})
	})
}

func testRun(t *testing.T, config ...experiment.ConfigFunc) {
	t.Run("it should record", func(t *testing.T) {
		tcs := map[string]struct {
			pubFunc func(context.Context, experiment.Observation[string]) error
		}{
			"a success": {
				pubFunc: func(_ context.Context, o experiment.Observation[string]) error {
					if o.Name == "correct" && !o.Success {
						t.Errorf("Expected a success, got mismatch")
					}

					return nil
				},
			},
			"a mismatch": {
				pubFunc: func(_ context.Context, o experiment.Observation[string]) error {
					if o.Name == "mismatch" && o.Success {
						t.Errorf("Expected a mismatch, got success")
					}

					return nil
				},
			},
			"an error": {
				pubFunc: func(_ context.Context, o experiment.Observation[string]) error {
					if o.Name == "error" && o.Error == nil {
						t.Errorf("Expected an error, got none")
					}

					return nil
				},
			},
			"the panic in the CandidatePanicError": {
				pubFunc: func(_ context.Context, o experiment.Observation[string]) error {
					if o.Name == "panic" {
						var panicError experiment.CandidatePanicError
						if !errors.As(o.Error, &panicError) {
							t.Errorf("Expected CandidatePanicError, got %T", o.Error)
						}

						if panicError.Panic == nil {
							t.Errorf("Expected a panic, did not record one")
						}
					}

					return nil
				},
			},
			"the clean control": {
				pubFunc: func(_ context.Context, o experiment.Observation[string]) error {
					if o.Name != "control" && o.Error == nil {
						if o.ControlValue != "Cleaned control" {
							t.Errorf("Expected value to be '%s', got '%s'", "Cleaned Control", o.ControlValue)
						}
					}

					return nil
				},
			},
			"the clean control value": {
				pubFunc: func(_ context.Context, o experiment.Observation[string]) error {
					if o.Error == nil && o.Success {
						cleaned := fmt.Sprintf("Cleaned %s", o.Value)
						if o.CleanValue != cleaned {
							t.Errorf("Expected value to be '%s', got '%s'", cleaned, o.CleanValue)
						}
					}

					return nil
				},
			},
		}

		for name, tc := range tcs {
			t.Run(name, func(t *testing.T) {
				ctx := context.Background()
				exp, pub := testExperiment(config...)

				var pubRun bool
				pub.fnc = func(ctx context.Context, o experiment.Observation[string]) error {
					pubRun = true
					return tc.pubFunc(ctx, o)
				}

				if _, err := exp.Run(ctx); err != nil {
					t.Errorf("Expected no error for running, got %s", err)
				}

				if err := exp.Publish(ctx); err != nil {
					t.Errorf("Expected no error for publishing, got %s", err)
				}

				if pubRun != true {
					t.Error("Expected publisher to run, it did not")
				}
			})
		}
	})

	t.Run("it should return the correct value", func(t *testing.T) {
		ctx := context.Background()
		exp, _ := testExperiment(config...)
		val, err := exp.Run(ctx)
		if val != "control" {
			t.Errorf("Expected value to be 'control', got '%s'", val)
		}

		if err != nil {
			t.Errorf("Expected error to be 'nil', got '%s'", err.Error())
		}
	})
}

func TestRun_Before(t *testing.T) {
	ctx := context.Background()
	exp := experiment.New[string]()
	exp.Force(true)

	exp.Control(func(context.Context) (string, error) {
		return "", nil
	})

	t.Run("with an error", func(t *testing.T) {
		expected := errors.New("before error")
		exp.Before(func(context.Context) error {
			return expected
		})

		_, err := exp.Run(ctx)
		if expected != err {
			t.Errorf("Expected error '%v', got '%v'", expected, err)
		}
	})

	t.Run("without an error", func(t *testing.T) {
		var ran bool
		exp.Before(func(context.Context) error {
			ran = true
			return nil
		})

		exp.Run(ctx)
		if !ran {
			t.Errorf("Expected before to have run, did not")
		}
	})
}

func TestRun_Ignore(t *testing.T) {
	exp, pub := testExperiment(experiment.WithPercentage(100))
	exp.Ignore(true)

	var count int
	var lock sync.Mutex
	pub.fnc = func(_ context.Context, _ experiment.Observation[string]) error {
		lock.Lock()
		defer lock.Unlock()
		count++
		return nil
	}

	exp.Run(context.Background())

	if count != 0 {
		t.Errorf("Expected '0' observations, got '%d'", count)
	}
}

func testWithTimeout(t *testing.T, o ...experiment.ConfigFunc) {
	contextTimeout := 10 * time.Millisecond
	timeFunc := func(ctx context.Context, dur time.Duration) error {
		tick := time.NewTicker(dur)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
			return nil
		}
	}

	t.Run("with slow control", func(t *testing.T) {
		exp := experiment.New[string](o...)
		exp.Force(true)

		var hasRun bool
		exp.Control(func(ctx context.Context) (string, error) {
			hasRun = true
			return "", timeFunc(ctx, time.Minute)
		})

		ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
		defer cancel()
		_, err := exp.Run(ctx)

		if !hasRun {
			t.Errorf("expected control to have run")
		}

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("expected deadline exceeded")
		}
	})

	t.Run("with slow candidate", func(t *testing.T) {
		pub := &testPublisher[string]{}
		exp := experiment.New[string](o...).WithPublisher(pub)
		exp.Force(true)

		var hasRun bool
		exp.Control(func(ctx context.Context) (string, error) {
			return "", nil
		})

		exp.Candidate("candidate", func(ctx context.Context) (string, error) {
			hasRun = true
			return "", timeFunc(ctx, time.Minute)
		})

		ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
		defer cancel()
		_, err := exp.Run(ctx)

		if !hasRun {
			t.Errorf("expected candidate to have run")
		}

		if err != nil {
			t.Errorf("expected no error from the control function, got %s", err)
		}

		// time.Sleep(time.Second)
		var candidateErr error
		pub.fnc = func(_ context.Context, o experiment.Observation[string]) error {
			if o.Name == "candidate" {
				candidateErr = o.Error
			}
			return nil
		}

		if err := exp.Publish(context.Background()); err != nil {
			t.Errorf("expected publishing to succeed, got error %s", err)
		}

		if !errors.Is(candidateErr, context.DeadlineExceeded) {
			t.Errorf("expected candidate to have exceeded the deadline, got %s error", candidateErr)
		}
	})

	t.Run("with everything timely", func(t *testing.T) {
		pub := &testPublisher[string]{}
		exp := experiment.New[string](o...).WithPublisher(pub)
		exp.Force(true)

		var hasRun bool
		exp.Control(func(ctx context.Context) (string, error) {
			return "", nil
		})

		exp.Candidate("candidate", func(ctx context.Context) (string, error) {
			hasRun = true
			return "", nil
		})

		ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
		defer cancel()
		_, err := exp.Run(ctx)

		if !hasRun {
			t.Errorf("expected candidate to have run")
		}

		if err != nil {
			t.Errorf("expected no error from the control function, got %s", err)
		}

		var candidateErr error
		pub.fnc = func(_ context.Context, o experiment.Observation[string]) error {
			if o.Name == "candidate" {
				candidateErr = o.Error
			}
			return nil
		}

		if err := exp.Publish(context.Background()); err != nil {
			t.Errorf("expected publishing to succeed, got error %s", err)
		}

		if candidateErr != nil {
			t.Errorf("expected no candidate error, got %s", candidateErr)
		}
	})
}

func TestRun_ContextPropagation(t *testing.T) {
	exp := experiment.New[string](experiment.WithTimeout(time.Second))
	exp.Force(true)

	key := func(s string) *string { return &s }("key")

	exp.Control(func(ctx context.Context) (string, error) {
		if value := ctx.Value(key); value != "value" {
			t.Errorf("expected context key to have value, got %s", value)
		}
		return "", nil
	})

	exp.Candidate("candidate", func(ctx context.Context) (string, error) {
		if value := ctx.Value(key); value != "value" {
			t.Errorf("expected context key to have value, got %s", value)
		}
		return "", nil
	})

	ctx := context.WithValue(context.Background(), key, "value")
	_, err := exp.Run(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
}

func TestRun_WithConcurrency(t *testing.T) {
	exp := experiment.New[string]()
	exp.Force(true)

	var count int
	var lock sync.Mutex

	expected := 5
	eChan := make(chan bool)

	fnc := func(name string) experiment.CandidateFunc[string] {
		return func(context.Context) (string, error) {
			lock.Lock()
			defer lock.Unlock()
			count++
			eChan <- true
			return name, nil
		}
	}

	exp.Control(fnc("control"))

	exp.Candidate("correct", fnc("correct"))

	exp.Candidate("mismatch", fnc("mismatch"))

	exp.Candidate("error", func(context.Context) (string, error) {
		lock.Lock()
		defer lock.Unlock()
		count++
		eChan <- true
		return "", errors.New("errored")
	})

	exp.Candidate("panic", func(context.Context) (string, error) {
		lock.Lock()
		defer lock.Unlock()
		count++
		eChan <- true
		panic("candidate")
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// run this in a goroutine so we can check for the messages coming in on the
	// eChan channel.
	go func() {
		if _, err := exp.Run(ctx); err != nil {
			t.Errorf("Expected no error running, got %s", err)
		}
	}()

checkLoop:
	for i := 0; i < expected; i++ {
		select {
		case <-eChan:
		case <-ctx.Done():
			t.Errorf("Expected to run the experiment within one second")
			break checkLoop
		}
	}

	if count != expected {
		t.Errorf("Expected '%d' observations, got '%d'", expected, count)
	}
}

func TestPublish_Errors(t *testing.T) {
	pub := &testPublisher[string]{}
	pub.fnc = func(ctx context.Context, o experiment.Observation[string]) error {
		return errors.New(o.Name)
	}

	exp := experiment.New[string]().WithPublisher(pub)
	exp.Force(true)

	exp.Control(func(context.Context) (string, error) {
		return "control", nil
	})

	exp.Candidate("correct", func(context.Context) (string, error) {
		return "control", nil
	})

	exp.Candidate("mismatch", func(context.Context) (string, error) {
		return "mismatch", nil
	})

	ctx := context.Background()
	if _, err := exp.Run(ctx); err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	var publishErr *experiment.PublishError
	if err := exp.Publish(ctx); !errors.As(err, &publishErr) {
		t.Errorf("Expected PublishError error, got %s", err)
	}

	if len := len(publishErr.Unwrap()); len != 3 {
		t.Errorf("Expected 3 errors, got %d", len)
	}
}

func testExperiment(cfg ...experiment.ConfigFunc) (*experiment.Experiment[string], *testPublisher[string]) {
	pub := &testPublisher[string]{}

	exp := experiment.New[string](cfg...).WithPublisher(pub)
	exp.Force(true)

	exp.Control(func(context.Context) (string, error) {
		return "control", nil
	})

	exp.Candidate("correct", func(context.Context) (string, error) {
		return "control", nil
	})

	exp.Candidate("mismatch", func(context.Context) (string, error) {
		return "mismatch", nil
	})

	exp.Candidate("error", func(context.Context) (string, error) {
		return "", errors.New("errored")
	})

	exp.Candidate("panic", func(context.Context) (string, error) {
		panic("candidate")
	})

	exp.Compare(func(control, candidate string) bool {
		return control == candidate
	})

	exp.Clean(func(c string) string {
		return fmt.Sprintf("Cleaned %s", c)
	})

	return exp, pub
}

type testPublisher[C any] struct {
	fnc func(context.Context, experiment.Observation[C]) error
}

func (t *testPublisher[C]) Publish(ctx context.Context, o experiment.Observation[C]) error {
	if t.fnc != nil {
		return t.fnc(ctx, o)
	}

	return nil
}
