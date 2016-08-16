package experiment_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jelmersnoeck/experiment"
)

func TestPublisher_Publish_PassedControl(t *testing.T) {
	p := &publisher{}
	pub := experiment.NewPublisher(p)
	obs := experiment.Observations{
		"control": newObservation("control", time.Minute),
	}

	pub.Publish(experiment.NewResult(obs, nil))
	if exp, val := 2, p.counts; exp != val {
		t.Fatalf("Expected counts to be `%d`, got `%d`", exp, val)
	}
	if exp, val := 1, p.timers; exp != val {
		t.Fatalf("Expected timers to be `%d`, got `%d`", exp, val)
	}
}

func TestPublisher_Publish_ErroredControl(t *testing.T) {
	p := &publisher{}
	pub := experiment.NewPublisher(p)
	obs := experiment.Observations{
		"control": newErrorObservation("control", time.Minute),
	}

	pub.Publish(experiment.NewResult(obs, nil))
	if exp, val := 2, p.increments; exp != val {
		t.Fatalf("Expected increments to be `%d`, got `%d`", exp, val)
	}
	if exp, val := 2, p.counts; exp != val {
		t.Fatalf("Expected counts to be `%d`, got `%d`", exp, val)
	}
	if exp, val := 1, p.timers; exp != val {
		t.Fatalf("Expected timers to be `%d`, got `%d`", exp, val)
	}
}

func TestPublisher_Publish_PassedTests(t *testing.T) {
	p := &publisher{}
	pub := experiment.NewPublisher(p)
	obs := experiment.Observations{
		"control": newObservation("control", time.Minute),
		"test":    newObservation("test", time.Minute),
	}

	pub.Publish(experiment.NewResult(obs, nil))
	if exp, val := 2, p.counts; exp != val {
		t.Fatalf("Expected counts to be `%d`, got `%d`", exp, val)
	}
	if exp, val := 2, p.timers; exp != val {
		t.Fatalf("Expected timers to be `%d`, got `%d`", exp, val)
	}
}

func TestPublisher_Publish_ErroredTests(t *testing.T) {
	p := &publisher{}
	pub := experiment.NewPublisher(p)
	obs := experiment.Observations{
		"control": newErrorObservation("control", time.Minute),
		"test":    newErrorObservation("test", time.Minute),
	}

	pub.Publish(experiment.NewResult(obs, nil))
	if exp, val := 3, p.increments; exp != val {
		t.Fatalf("Expected increments to be `%d`, got `%d`", exp, val)
	}
	if exp, val := 2, p.counts; exp != val {
		t.Fatalf("Expected counts to be `%d`, got `%d`", exp, val)
	}
	if exp, val := 2, p.timers; exp != val {
		t.Fatalf("Expected timers to be `%d`, got `%d`", exp, val)
	}
}

func BenchmarkPublisher_Publish(b *testing.B) {
	pub := experiment.NewPublisher(&publisher{})
	obs := experiment.Observations{
		"control": newObservation("control", time.Minute),
		"test":    newObservation("test", time.Minute),
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pub.Publish(experiment.NewResult(obs, nil))
		}
	})
}

func newObservation(name string, duration time.Duration) experiment.Observation {
	return experiment.Observation{
		Name:     name,
		Duration: duration,
	}
}

func newErrorObservation(name string, duration time.Duration) experiment.Observation {
	return experiment.Observation{
		Name:     name,
		Duration: duration,
		Error:    errors.New(name),
	}
}

type publisher struct {
	sync.Mutex
	counts     int
	increments int
	timers     int
}

func (p *publisher) Increment(key string) {
	p.Lock()
	defer p.Unlock()

	p.increments++
}

func (p *publisher) Count(key string, i interface{}) {
	p.Lock()
	defer p.Unlock()

	p.counts++
}

func (p *publisher) Timing(key string, i interface{}) {
	p.Lock()
	defer p.Unlock()

	p.timers++
}
