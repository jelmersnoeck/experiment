package experiment

// Config represents the configuration options for an experiment.
type Config struct {
	Percentage  int
	Concurrency bool
}

// ConfigFunc represents a function that knows how to set a configuration option.
type ConfigFunc func(*Config)

// Publisher represents an interface that allows you to publish results.
type Publisher[C any] interface {
	Publish(Observation[C])
}

// WithPercentage returns a new func(*Config) that sets the percentage.
func WithPercentage(p int) ConfigFunc {
	return func(c *Config) {
		c.Percentage = p
	}
}

// WithConcurrency forces the experiment to run concurrently
func WithConcurrency() ConfigFunc {
	return func(c *Config) {
		c.Concurrency = true
	}
}

// WithDefaultConfig returns a new configuration with defaults.
func WithDefaultConfig() ConfigFunc {
	return func(c *Config) {
		c.Percentage = 0
		c.Concurrency = false
	}
}
