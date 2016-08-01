package experiment

// ResultPublisher is what we'll use to make our results visible. Multiple
// publishers can be used for a single experiment with each having their own
// purpose.
type ResultPublisher interface {
	Publish(Result)
}
