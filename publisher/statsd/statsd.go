package statsd

import (
	"fmt"
	"time"

	"github.com/alexcesaro/statsd"
	"github.com/jelmersnoeck/experiment"
)

type statsdPublisher struct {
	pf string
	cl *statsd.Client
}

// New creates a new ResultPublisher that will publish results to a statsd
// client.
func New(prefix string, opts ...statsd.Option) (experiment.ResultPublisher, error) {
	cl, err := statsd.New(opts...)
	if err != nil {
		return nil, err
	}

	return &statsdPublisher{pf: prefix, cl: cl}, nil
}

func (p *statsdPublisher) Publish(res experiment.Result) {
	p.publishObservation(res.Control())
	for _, ob := range res.Candidates() {
		p.publishObservation(ob)
	}

	p.cl.Count(p.bucketName("mismatches")+".count", len(res.Mismatches()))
}

func (p *statsdPublisher) publishObservation(ob experiment.Observation) {
	if ob.Error != nil {
		p.cl.Increment(p.bucketName(ob.Name) + ".errors.incr")
	}

	if ob.Panic != nil {
		p.cl.Increment(p.bucketName(ob.Name) + ".panics.incr")
	}

	p.cl.Timing(
		p.bucketName(ob.Name)+".time",
		int(ob.Duration/time.Millisecond),
	)
}

func (p *statsdPublisher) bucketName(name string) string {
	return fmt.Sprintf("%s.%s", p.pf, name)
}
