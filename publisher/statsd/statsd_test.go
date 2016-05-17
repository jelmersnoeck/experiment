package statsd_test

import (
	"github.com/jelmersnoeck/experiment"
	"github.com/jelmersnoeck/experiment/publisher/statsd"
)

func ExampleNew() {
	exp := experiment.New("test")
	res, _ := exp.Result()

	pub, _ := statsd.New(exp)
	pub.Publish(res)
}
