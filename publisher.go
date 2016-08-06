package experiment

import "time"

type (
	// ResultPublisher is what we'll use to make our results visible. Multiple
	// publishers can be used for a single experiment with each having their own
	// purpose.
	Publisher interface {
		Publish(Result)
	}

	// PublisherClient takes care of actually sending data to a collector to
	// visualise your data. This could be a statsd or graphite client for example.
	PublisherClient interface {
		Increment(string)
		Count(string, interface{})
		Timing(string, interface{})
	}

	publisher struct {
		cl PublisherClient
	}
)

// New creates a new ResultPublisher that will publish results to a publisher
// client.
func NewPublisher(client PublisherClient) Publisher {
	return &publisher{cl: client}
}

func (p *publisher) Publish(res Result) {
	p.publishObservation(res.Control())
	for _, ob := range res.Candidates() {
		p.publishObservation(ob)
	}

	p.cl.Count("candidates.count", len(res.Candidates()))
	p.cl.Count("mismatches.count", len(res.Mismatches()))
}

func (p *publisher) publishObservation(ob Observation) {
	if ob.Error != nil {
		p.cl.Increment(ob.Name + ".errors.incr")
	}

	if ob.Panic != nil {
		p.cl.Increment(ob.Name + ".panics.incr")
	}

	p.cl.Timing(
		ob.Name+".time",
		int(ob.Duration/time.Millisecond),
	)
}
