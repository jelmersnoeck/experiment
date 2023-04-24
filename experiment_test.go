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
		pub.fnc = func(o experiment.Observation[string]) {
			if o.Name == "correct" && !o.Success {
				t.Errorf("Expected a success, got mismatch")
			}
		}

		exp.Run()
	})

	t.Run("it should record a mismatch", func(t *testing.T) {
		pub.fnc = func(o experiment.Observation[string]) {
			if o.Name == "mismatch" && o.Success {
				t.Errorf("Expected a mismatch, got success")
			}
		}

		exp.Run()
	})

	t.Run("it should record an error", func(t *testing.T) {
		pub.fnc = func(o experiment.Observation[string]) {
			if o.Name == "error" && o.Error == nil {
				t.Errorf("Expected an error, got none")
			}
		}

		exp.Run()
	})

	t.Run("it should record the panic", func(t *testing.T) {
		pub.fnc = func(o experiment.Observation[string]) {
			if o.Name == "panic" && o.Panic == nil {
				t.Errorf("Expected a panic, did not record one")
			}
		}

		exp.Run()
	})

	t.Run("it should record the clean control", func(t *testing.T) {
		pub.fnc = func(o experiment.Observation[string]) {
			if o.Name != "control" && o.Error == nil {
				if o.Panic == nil && o.ControlValue != "Cleaned control" {
					t.Errorf("Expected value to be '%s', got '%s'", "Cleaned Control", o.ControlValue)
				}
			}
		}

		exp.Run()
	})

	t.Run("it should record the clean control value", func(t *testing.T) {
		pub.fnc = func(o experiment.Observation[string]) {
			if o.Panic == nil && o.Error == nil && o.Success {
				cleaned := fmt.Sprintf("Cleaned %s", o.Value)
				if o.CleanValue != cleaned {
					t.Errorf("Expected value to be '%s', got '%s'", cleaned, o.CleanValue)
				}
			}
		}

		exp.Run()
	})

	t.Run("it should return the correct value", func(t *testing.T) {
		val, err := exp.Run()
		if val != "control" {
			t.Errorf("Expected value to be 'control', got '%s'", val)
		}

		if err != nil {
			t.Errorf("Expected error to be 'nil', got '%s'", err.Error())
		}
	})
}

func TestRun_Before(t *testing.T) {
	exp := experiment.New[string]()
	exp.Force(true)

	exp.Control(func() (string, error) {
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
	pub.fnc = func(_ experiment.Observation[string]) {
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	exp, pub := testExperiment(experiment.WithConcurrency())
	expected := 5
	eChan := make(chan bool)

	var count int
	var lock sync.Mutex
	pub.fnc = func(o experiment.Observation[string]) {
		lock.Lock()
		defer lock.Unlock()
		count++
		eChan <- true
	}

	exp.Run()

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

func testExperiment(cfg ...experiment.ConfigFunc) (*experiment.Experiment[string], *testPublisher[string]) {
	pub := &testPublisher[string]{}

	exp := experiment.New[string](cfg...).WithPublisher(pub)
	exp.Force(true)

	exp.Control(func() (string, error) {
		return "control", nil
	})

	exp.Candidate("correct", func() (string, error) {
		return "control", nil
	})

	exp.Candidate("mismatch", func() (string, error) {
		return "mismatch", nil
	})

	exp.Candidate("error", func() (string, error) {
		return "", errors.New("errored")
	})

	exp.Candidate("panic", func() (string, error) {
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
	fnc func(experiment.Observation[C])
}

func (t *testPublisher[C]) Publish(o experiment.Observation[C]) {
	if t.fnc != nil {
		t.fnc(o)
	}
}
