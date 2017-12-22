# Experiment

[Examples](_examples) | [Contributing](CONTRIBUTING.md) | [Code of Conduct](.github/CODE_OF_CONDUCT.md) | [License](LICENSE)

[![GitHub release](https://img.shields.io/github/tag/jelmersnoeck/experiment.svg?label=latest)](https://github.com/jelmersnoeck/experiment/releases)
[![Build Status](https://travis-ci.org/jelmersnoeck/experiment.svg?branch=master)](https://travis-ci.org/jelmersnoeck/experiment)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![GoDoc](https://godoc.org/github.com/jelmersnoeck/experiment?status.svg)](https://godoc.org/github.com/jelmersnoeck/experiment)
[![Report Card](https://goreportcard.com/badge/github.com/jelmersnoeck/experiment)](https://goreportcard.com/report/github.com/jelmersnoeck/experiment)
[![codecov](https://codecov.io/gh/jelmersnoeck/experiment/branch/master/graph/badge.svg)](https://codecov.io/gh/jelmersnoeck/experiment)

Experiment is a Go package to test and evaluate new code paths without
interfering with the users end result.

This is inspired by the [GitHub Scientist gem](https://github.com/github/scientist).

## Usage

### Control

`Control(func() (interface{}, error))` should be used to implement your current
code. The result of this will be used to compare to other candidates. This will
run as it would run normally.

A control is always expected. If no control is provided, the experiment will
panic.

```go
func main() {
	exp := experiment.New(
		experiment.WithPercentage(50),
	)

	exp.Control(func() (interface{}, error) {
		return fmt.Sprintf("Hello world!"), nil
	})

	result, err := exp.Run()
	if err != nil {
		panic(err)
	} else {
		fmt.Println(result.(string))
	}
}
```

The example above will always print `Hello world!`.

### Candidate

`Candidate(string, func() (interface{}, error))` is a potential refactored
candidate. This will run sandboxed, meaning that when this panics, the panic
is captured and the experiment continues.

A candidate will not always run, this depends on the `WithPercentage(int)`
configuration option and further overrides.

```go
func main() {
	exp := experiment.New(
		experiment.WithPercentage(50),
	)

	exp.Control(func() (interface{}, error) {
		return fmt.Sprintf("Hello world!"), nil
	})

	exp.Candidate("candidate1", func() (interface{}, error) {
		return "Hello candidate", nil
	})

	result, err := exp.Run()
	if err != nil {
		panic(err)
	} else {
		fmt.Println(result.(string))
	}
}
```

The example above will still only print `Hello world!`. The `candidate1`
function will however run in the background 50% of the time.

### Run

`Run()` will run the experiment and return the value and error of the control
function. The control function is always executed. The result value of the
`Run()` function is an interface. The user should cast this to the expected
type.

### Force

`Force(bool)` allows you to force run an experiment and overrules all other
options. This can be used in combination with feature flags or to always run
the experiment for admins for example.

### Ignore

`Ignore(bool)` will disable the experiment, meaning that it will only run the
control function, nothing else.

### Compare

`Compare(interface{}, interface{}) bool` is used to compare the control value
against a candidate value.

If the candidate returned an error, this will not be executed.

### Clean

`Clean(interface{}) interface{}` is used to clean the output values. This is
implemented so that the publisher could use this cleaned data to store for later
usage.

If the candidate returned an error, this will not be executed and the
`CleanValue` field will be populated by the original `Value`.

## Limitations and caveats

### Stateless

Due to the fact that it is not guaranteed that a test will run every time or in
what order a test will run, it is suggested that experiments only do stateless
changes.

When enabling the `WithConcurrency()` option, keep in mind that your tests will
run concurrently in a random fashion. Make sure accessing your data concurrently
is allowed.

### Performance

By default, the candidates run sequentially. This means that there could be a
significant performance degradation due to slow new functionality.

### Memory leaks

When running with the `WithConcurrency()` option, the tests will run
concurrently and the control result will be returned as soon as possible. This
does however mean that the other candidates are still running in the background.
Be aware that this could lead to potential memory leaks and should thus be
monitored closely.

## Observation

An Observation contains several attributes. The first one is the `Value`. This
is the value which is returned by the control function that is specified. There
is also an `Error` attribute available, which contains the error returned.

## Errors

### Regular errors

When the control errors, this will be returned in the `Run()` method. When a
candidate errors, this will be attached to the `Error` field in its observation.

### Panics

When the control panics, this panic will be respected and actually be triggered.
When a candidate function panics, the experiment will swallow this and assign
this to the `Panic` field of the observation, which you can use in the
Publisher. An `ErrCandidatePanic` will also be returned.

## Config

### WithConcurrency()

If the `WithConcurrency()` configuration option is passed to the constructor,
the experiment will run its candidates in parallel. The result of the control
will be returned as soon as it's finished. Other work will continue in the
background.

This is disabled by default.

### WithPercentage(int)

`WithPercentage(int)` allows you to set the amount of time you want to run the
experiment as a percentage. `Force` and `Ignore` do not have an impact on this.

This is set to 0 by default to encourage setting a sensible percentage.

### WithPublisher(Publisher)

`WithPublisher(Publisher)` marks the experiment as Publishable. This means that
all the results will be pushed to the Publisher once the experiment has run.

This is nil by default.
