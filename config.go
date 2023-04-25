package experiment

import "time"

// Config represents the configuration options for an experiment.
type Config struct {
	Percentage  int
	Concurrency bool
	Timeout     *time.Duration
}

// ConfigFunc represents a function that knows how to set a configuration option.
type ConfigFunc func(*Config)

// WithPercentage returns a new func(*Config) that sets the percentage.
func WithPercentage(p int) ConfigFunc {
	return func(c *Config) {
		c.Percentage = p
	}
}

func WithTimeout(t time.Duration) ConfigFunc {
	return func(c *Config) {
		c.Timeout = &t
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
