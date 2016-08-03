package statsd_test

import (
	"testing"
	"time"

	astatsd "github.com/alexcesaro/statsd"
	"github.com/jelmersnoeck/experiment"
	"github.com/jelmersnoeck/experiment/publisher/statsd"
	"github.com/stretchr/testify/require"
)

func ExampleNew() {
	res := experiment.NewResult(experiment.Observations{}, nil)
	pub, _ := statsd.New("my-experiment")
	pub.Publish(res)
}

func BenchmarkStatsd_Publish(b *testing.B) {
	pub, err := statsd.New("my-experiment", astatsd.Mute(true))
	require.Nil(b, err)
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
