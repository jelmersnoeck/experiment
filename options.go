package experiment

type (
	options struct {
		name       string
		enabled    bool
		percentage float64
		comparison ComparisonMethod
	}

	ComparisonMethod func(Observation, Observation) bool
	Option           func(*options)
)

func newOptions(ops ...Option) options {
	opts := options{
		enabled:    true,
		percentage: 10,
	}

	for _, o := range ops {
		o(&opts)
	}

	return opts
}

// Name sets the name of the experiment. This will be used to build a report. If
// no name is given as an option to `New()`, the `NoNameError` will be returned.
func Name(name string) Option {
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
