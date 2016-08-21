package experiment

import "time"

type (
	// PublisherClient takes care of actually sending data to a collector to
	// visualise your data. This could be a statsd or graphite client for example.
	PublisherClient interface {
		Increment(string)
		Count(string, interface{})
		Timing(string, interface{})
	}

	// Publisher is what we'll use to make our results visible. Multiple
	// Publishers can be used for a single experiment with each having their own
	// purpose.
	Publisher struct {
		cl PublisherClient
	}
)

// NewPublisher creates a new publisher that will publish results to a client.
func NewPublisher(client PublisherClient) *Publisher {
	return &Publisher{cl: client}
}

// Publish takes a resultset and sends it off to the client in the publisher.
// This will publish the number of candidates, mismatches and a hit counter.
// Per observation - including the control - it will also publish the error
// count, panic count and observation duration.
func (p *Publisher) Publish(res *Result) {
	p.publishObservation(res.Control())
	for _, ob := range res.Candidates() {
		p.publishObservation(ob)
	}

	p.cl.Increment("publish.incr")
	p.cl.Count("candidates.count", len(res.Candidates()))
	p.cl.Count("mismatches.count", len(res.Mismatches()))
}

func (p *Publisher) publishObservation(ob Observation) {
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
