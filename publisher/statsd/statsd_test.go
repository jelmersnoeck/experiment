package statsd_test

import (
	"github.com/jelmersnoeck/experiment"
	"github.com/jelmersnoeck/experiment/publisher/statsd"
)

func ExampleNew() {
	res := experiment.NewResult(experiment.Observations{}, nil)
	pub, _ := statsd.New("my-experiment")
	pub.Publish(res)
}
