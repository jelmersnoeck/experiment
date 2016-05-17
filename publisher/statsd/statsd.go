package statsd

import (
	"fmt"

	"github.com/alexcesaro/statsd"
	"github.com/jelmersnoeck/experiment"
)

type statsdPublisher struct {
	cl  *statsd.Client
	exp *experiment.Experiment
}

// New creates a new ResultPublisher that will publish results to a statsd
// client.
func New(exp *experiment.Experiment, opts ...statsd.Option) (experiment.ResultPublisher, error) {
	cl, err := statsd.New(opts...)
	if err != nil {
		return nil, err
	}

	return &statsdPublisher{cl, exp}, nil
}

func (p *statsdPublisher) Publish(res experiment.Result) {
	p.publishObservation(res.Control())
	for _, ob := range res.Candidates() {
		p.publishObservation(ob)
	}
}

func (p *statsdPublisher) publishObservation(ob experiment.Observation) {
	p.cl.Timing(p.bucketName(ob.Name()), ob.Duration())
}

func (p *statsdPublisher) bucketName(name string) string {
	return fmt.Sprintf("%s.%s", p.exp.Name(), name)
}
