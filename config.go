package experiment

// Config represents the configuration options for an experiment.
type Config struct {
	Percentage  int
	Publisher   Publisher
	Concurrency bool
}

// Publisher represents an interface that allows you to publish results.
type Publisher interface {
	Publish(Observation)
}

// ConfigFunc represents a function that knows how to set a configuration option.
type ConfigFunc func(*Config)

// WithPercentage returns a new ConfigFunc that sets the percentage.
func WithPercentage(p int) ConfigFunc {
	return func(c *Config) {
		c.Percentage = p
	}
}

// WithPublisher returns a new ConfigFunc that sets the publisher.
func WithPublisher(p Publisher) ConfigFunc {
	return func(c *Config) {
		c.Publisher = p
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
		c.Publisher = nil
	}
}
