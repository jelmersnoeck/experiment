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

func TestRun_Sequential(t *testing.T) {
	exp, pub := testExperiment()

	t.Run("it should record a success", func(t *testing.T) {
		pub.fnc = func(o experiment.Observation) {
			if o.Name == "correct" && !o.Success {
				t.Errorf("Expected a success, got mismatch")
			}
		}

		exp.Run()
	})

	t.Run("it should record a mismatch", func(t *testing.T) {
		pub.fnc = func(o experiment.Observation) {
			if o.Name == "mismatch" && o.Success {
				t.Errorf("Expected a mismatch, got success")
			}
		}

		exp.Run()
	})

	t.Run("it should record an error", func(t *testing.T) {
		pub.fnc = func(o experiment.Observation) {
			if o.Name == "error" && o.Error == nil {
				t.Errorf("Expected an error, got none")
			}
		}

		exp.Run()
	})

	t.Run("it should record the panic", func(t *testing.T) {
		pub.fnc = func(o experiment.Observation) {
			if o.Name == "panic" && o.Panic == nil {
				t.Errorf("Expected a panic, did not record one")
			}
		}

		exp.Run()
	})

	t.Run("it should record the clean", func(t *testing.T) {
		pub.fnc = func(o experiment.Observation) {
			if o.Panic == nil && o.Error == nil && o.Success {
				cleaned := fmt.Sprintf("Cleaned %s", o.Value.(string))
				if o.CleanValue.(string) != cleaned {
					t.Errorf("Expected value to be '%s', got '%s'", cleaned, o.CleanValue.(string))
				}
			}
		}

		exp.Run()
	})

	t.Run("it should return the correct value", func(t *testing.T) {
		val, err := exp.Run()
		if val.(string) != "control" {
			t.Errorf("Expected value to be 'control', got '%s'", val.(string))
		}

		if err != nil {
			t.Errorf("Expected error to be 'nil', got '%s'", err.Error())
		}
	})
}

func TestRun_Before(t *testing.T) {
	exp := experiment.New()
	exp.Force(true)

	exp.Control(func() (interface{}, error) {
		return "", nil
	})

	t.Run("with an error", func(t *testing.T) {
		expected := errors.New("before error")
		exp.Before(func() error {
			return expected
		})

		_, err := exp.Run()
		if expected != err {
			t.Errorf("Expected error '%v', got '%v'", expected, err)
		}
	})

	t.Run("without an error", func(t *testing.T) {
		var ran bool
		exp.Before(func() error {
			ran = true
			return nil
		})

		exp.Run()
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
	pub.fnc = func(_ experiment.Observation) {
		lock.Lock()
		defer lock.Unlock()
		count++
	}

	exp.Run()

	if count != 0 {
		t.Errorf("Expected '0' observations, got '%d'", count)
	}
}

func TestRun_Concurrent(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	exp, pub := testExperiment(experiment.WithConcurrency())
	expected := 5
	eChan := make(chan bool)

	var count int
	var lock sync.Mutex
	pub.fnc = func(_ experiment.Observation) {
		lock.Lock()
		defer lock.Unlock()
		count++
		eChan <- true
	}

	exp.Run()

	for i := 0; i < expected; i++ {
		select {
		case <-eChan:
		case <-ctx.Done():
			t.Errorf("Expected to run the experiment within one second")
			break
		}
	}

	if count != expected {
		t.Errorf("Expected '%d' observations, got '%d'", expected, count)
	}
}

func testExperiment(cfg ...experiment.ConfigFunc) (*experiment.Experiment, *testPublisher) {
	pub := &testPublisher{}

	config := []experiment.ConfigFunc{
		experiment.WithPublisher(pub),
	}
	config = append(config, cfg...)

	exp := experiment.New(config...)
	exp.Force(true)

	exp.Control(func() (interface{}, error) {
		return "control", nil
	})

	exp.Candidate("correct", func() (interface{}, error) {
		return "control", nil
	})

	exp.Candidate("mismatch", func() (interface{}, error) {
		return "mismatch", nil
	})

	exp.Candidate("error", func() (interface{}, error) {
		return nil, errors.New("errored")
	})

	exp.Candidate("panic", func() (interface{}, error) {
		panic("candidate")
	})

	exp.Compare(func(control interface{}, candidate interface{}) bool {
		return control.(string) == candidate.(string)
	})

	exp.Clean(func(c interface{}) interface{} {
		return fmt.Sprintf("Cleaned %s", c.(string))
	})

	return exp, pub

}

type testPublisher struct {
	fnc func(experiment.Observation)
}

func (t *testPublisher) Publish(o experiment.Observation) {
	if t.fnc != nil {
		t.fnc(o)
	}
}
