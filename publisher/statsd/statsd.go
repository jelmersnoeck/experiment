package statsd

import (
	"fmt"
	"time"

	"github.com/alexcesaro/statsd"
	"github.com/jelmersnoeck/experiment"
)

type statsdPublisher struct {
	cl  *statsd.Client
	exp *experiment.Experiment
}

// New creates a new ResultPublisher that will publish results to a statsd
// client.
func New(opts ...statsd.Option) (experiment.ResultPublisher, error) {
	cl, err := statsd.New(opts...)
	if err != nil {
		return nil, err
	}

	return &statsdPublisher{cl: cl}, nil
}

func (p *statsdPublisher) Publish(exp *experiment.Experiment, res experiment.Result) {
	p.exp = exp
	p.publishObservation(res.Control())
	for _, ob := range res.Candidates() {
		p.publishObservation(ob)
	}
}

func (p *statsdPublisher) publishObservation(ob experiment.Observation) {
	p.cl.Timing(
		p.bucketName(ob.Name()),
		ob.Duration().Nanoseconds()*time.Millisecond,
	)
}

func (p *statsdPublisher) bucketName(name string) string {
	return fmt.Sprintf("%s.%s", p.exp.Name(), name)
}
