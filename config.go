package experiment

import "golang.org/x/net/context"

var (
	// TestMode indicates if the code is executed as part of the test suite.
	// When TestMode is enabled, all the experiment tests will always run,
	// regardless of any other settings. Any potential panics that are caused
	// in any of the tests will also not be recovered but actually panic.
	TestMode = false
)

type (
	// Config is the configuration used to set up an experiment. Config is not
	// safe for concurrent use.
	Config struct {
		Percentage    float32 `json:"percentage"`
		BeforeFilters []BeforeFilter
	}

	// BeforeFilter is a wrapper around a method that is purely used to take a
	// context, adjust it and return a new context with the adjusted values.
	BeforeFilter func(context.Context) context.Context

	// ComparisonMethod is used as an interface for creating a method in which we
	// want to compare the observations of a test. This being the Control and a
	// random Test case.
	ComparisonMethod func(Observation, Observation) bool

	// ConditionalFunc is used to determine on run time whether or not we should
	// run the tests.
	ConditionalFunc func(context.Context) bool
)

// DefaultConfig sets up a default configuration where the Percentage is 100.
func DefaultConfig() Config {
	return Config{
		Percentage: 100,
	}
}
