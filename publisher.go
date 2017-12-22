package experiment

import "log"

// Logger represents the interface that experiment expects for a logger.
type Logger interface {
	Printf(string, ...interface{})
}

// NewLogPublisher returns a new LogPublisher.
func NewLogPublisher(name string, logger Logger) *LogPublisher {
	return &LogPublisher{
		Name:   name,
		Logger: logger,
	}
}

// LogPublisher is a publisher that writes out the observation values as a log
// line. If no Logger is provided, the standard library logger will be used.
type LogPublisher struct {
	Name   string
	Logger Logger
}

// Publish will publish all the Observation variables as a log line. It is in
// the following format:
// [Experiment Observation] name=%s duration=%s success=%t value=%v error=%v
func (l *LogPublisher) Publish(o Observation) {
	msg := "[Experiment Observation: %s] name=%s duration=%s success=%t value=%v error=%v"
	args := []interface{}{l.Name, o.Name, o.Duration, o.Success, o.CleanValue, o.Error}
	if l.Logger == nil {
		log.Printf(msg, args...)
	} else {
		l.Logger.Printf(msg, args...)
	}
}
