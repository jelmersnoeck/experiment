package experiment

import "golang.org/x/net/context"

type (
	options struct {
		name       string
		enabled    bool
		testMode   bool
		percentage float64
		comparison ComparisonMethod
		ctx        context.Context
		before     []ContextMethod
		publishers []ResultPublisher
	}

	ComparisonMethod func(Observation, Observation) bool
	ContextMethod    func(context.Context) context.Context
	Option           func(*options)
)

func newOptions(ops ...Option) options {
	opts := options{
		enabled:    true,
		percentage: 10,
		ctx:        context.Background(),
		before:     []ContextMethod{},
		publishers: []ResultPublisher{},
	}

	for _, o := range ops {
		o(&opts)
	}

	return opts
}

func name(name string) Option {
	return func(opts *options) {
		opts.name = name
	}
}

// Percentage sets the percentage on how many times we should run the test.
// Internally, we'll keep a counter for each experiment and based on that we'll
// decide if the experiment should actually run when calling the `Run` method.
func Percentage(p int) Option {
	return func(opts *options) {
		opts.percentage = float64(p)
	}
}

// Enabled is basically a conditional around the experiment. The reason this is
// included is to have a consistent way in your code to define experiments
// without having to wrap them in conditionals. This way, you can create a
// minimalistic check and pass it to the experiment and write code as if the
// experiment is enabled.
func Enabled(b bool) Option {
	return func(opts *options) {
		opts.enabled = b
	}
}

// Compare is the method that is used to compare the results from the test. The
// control and test function should always return a value. These values will
// then be injected in the compare method. When we publish the results for this
// test, we will use the value from this compare method to look at the success
// rate of our test.
func Compare(m ComparisonMethod) Option {
	return func(opts *options) {
		opts.comparison = m
	}
}

// TestMode is used to set the experiment runner in test mode. This means that
// the tests will always be run, no matter what other options are given. This
// also means that any potential panics will occur instead of being ignored.
func TestMode() Option {
	return func(opts *options) {
		opts.testMode = true
	}
}

// Context is an option that allows you to add a context to the experiment. This
// will be used as a base for injecting the context into your test methods.
func Context(ctx context.Context) Option {
	return func(opts *options) {
		opts.ctx = ctx
	}
}

// Before allows someone to set a setup operation that could be an expensive
// task, and that shouldn't be executed every time. Context is injected in this
// method so that the data that has been setup can be used at a later time.
func Before(bef ContextMethod) Option {
	return func(opts *options) {
		opts.before = append(opts.before, bef)
	}
}

// Publisher adds a publisher to our setup. This will be used to publish our
// experiment results to. Multiple publishers can be configured per experiment.
// For example, a statsd publisher can be used to measure timings and a redis
// publisher can be used to store comparison data.
func Publisher(p ResultPublisher) Option {
	return func(opts *options) {
		opts.publishers = append(opts.publishers, p)
	}
}
