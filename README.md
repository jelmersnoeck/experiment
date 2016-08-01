# Experiment

[![Build Status](https://travis-ci.org/jelmersnoeck/experiment.svg?branch=master)](https://travis-ci.org/jelmersnoeck/experiment)
[![GoDoc](https://godoc.org/github.com/jelmersnoeck/experiment?status.svg)](https://godoc.org/github.com/jelmersnoeck/experiment)

Experiment is a Go package to test and evaluate new code paths without
interfering with the users end result.

This is inspired by the [GitHub Scientist gem](https://github.com/github/scientist).

## Usage

Below is the most basic example on how to create an experiment. We'll see if
using a string with the notation of `""` compares to using a byte buffer and
we'll evaluate the time difference.

```go
func main() {
	exp := experiment.New(experiment.DefaultConfig("my-test"))

	exp.Control(func(ctx context.Context) (interface{}, error) {
		return "my-text", nil
	})
	exp.Test("buffer", func(ctx context.Context) (interface{}, error) {
		buf := bytes.NewBufferString("")
		buf.WriteString("new")
		buf.Write([]byte(`-`))
		buf.WriteString("text")

		return string(buf.Bytes()), nil
	})

	obs, err := exp.Run(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	str = obs.Control().Value.(string)
	fmt.Println(str)
}
```

First, we create an experiment with a new name. This will identify the
experiment later on in our publishers.

Further down, we set a control function. This is basically the functionality you
are currently using in your codebase and want to evaluate against. The `Control`
method is of the same structure as the `Test` method, in which it takes a
`context.Context` to pass along any data and expects an interface and error as
return value.

The next step is to define tests. These tests are to see if the newly refactored
code performs better and yields the same output. The sampling rate for these
tests can be configured as mentioned later on in the options section.

Once the setup is complete, it is time to run our experiment. This will run our
control and tests(if applicable) in a random fashion. This means that one time
the control could be run first, another time the test case could be run first.
This is done so to avoid any accidental behavioural changes in any of the
control or test code.

The `Run()` method also returns an `Observations` type. This is a set of
observations which includes the control observation. This can be accessed by
calling the `Control()` method on the `Observations`. It also contains the
observations for all the other test cases if run.

If an error happens trying to run the experiment, this will be returned. If an
error happens when running a test, this will be available on the `Observation`
by accessing the `Error()` method.

## Limitations

### Stateless

Due to the fact that it is not guaranteed that a test will run every time or in
what order a test will run, it is suggested that experiments only do stateless
changes.

### Multiple tests

Although it is possible to add multiple test cases to a single experiment, it is
not suggested to do so. The test are run synchroniously which means this can add
up to your response time.

## Run

The `Run()` method executes the experiment. This means that it will run the
control and potentially the tests.

The control will be run no matter what. The tests might run depending on several
options (see listed below).

## ForceRun

The `ForceRun()` method has the same implementation as the `Run()` method apart
from the conditional run checks. This means that the tests will always run no
matter what the other options are.

## Observation

An Observation contains several attributes. The first one is the `Value`. This
is the value which is returned by the control function that is specified. There
is also an `Error` attribute available, which contains the error returned.

## Panics

When the control panics, this panic will be respected and actually be triggered.
When a test function panics, the experiment will swallow this and add the panic
to the `Panic` attribute on the `Observation`.

## Config

When creating a new experiment, one can add several options. Some of them have
default values.

### Comparison

By default, the duration and return values will be captured. If you want to
conclude the experiment by comparing results, a `Compare` option can be given
when creating a new experiment. Mismatches will then be recorded in the `Result`
which can be used to be published.

If no `Compare` option is given, no mismatches will be recorded, only durations.

```go
func main() {
    cfg := experiment.Config{
        Name: "comparison",
        Percentage: 10,
        Comparison: comparisonMethod,
    }
	exp := experiment.New(cfg)

	// add control/tests

	obs, err := exp.Run(nil)
    if err != nil {
        fmt.Println(err)
        return
    }

    res := experiment.NewResult(obs, cfg)
    fmt.Println(res.Mismatches())
}

func comparisonMethod(control experiment.Observation, test experiment.Observation) bool {
	c := control.Value().(string)
	t := test.Value().(string)

	return c == t
}
```

With a `Compare` option set, the result generator (called by `Result()`) will go
over all observations and call this comparison method with both the control
observation and the test case observation. It will then store all mismatched
separately in the `Mismatches()` method on the `Result` object.

### Percentage

Sometimes, you don't want to run the experiment for every request. To do this,
one can set the percentage rate of which we'll run our test cases.

The control will always be run when using the `Run()` method. This option just
disables the test cases to be run.


```go
func main() {
    cfg := experiment.Config{
        Name: "percentage",
        Percentage: 25,
    }
	exp := experiment.New(cfg)

	// add control/tests

	obs, err := exp.Run(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	str = obs.Control().Value.(string)
	fmt.Println(str)
}
```

Now that we've set the percentage to 25, the experiment will only be run 1/4
times. This is good for sampling data and rolling it out sequentially.

### Before

When an expensive setup is required to do the test, we don't always want to run
the setup until we actually execute the test case. The `Before` option allows us
to set this up. This means that when an experiment is not run, this set up will
not be executed.

To do this, we make use of context.

```go
func main() {
    cfg := experiment.Config{
        Name: "percentage",
        Percentage: 25,
        BeforeFilters: []BeforeFilter{mySetup},
    }
	exp := experiment.New(cfg)

	// do more experiment setup and run it
}

func mySetup(ctx context.Context) context.Context {
	expensive := myExpensiveSetup()
	return context.WithValue(ctx, "my-thing", expensive)
}

func myControlFunc(ctx context.Context) (interface{}, error) {
	thing := ctx.Value("my-thing")

	// logic with `thing`
}
```

In the above example, we create a new expensive setup somehow (this is not
implemented in the example). We then pass this function to our new experiment.

If the experiment runner decides to run the experiment - based on percentage and
other options - the setup will be executed. If not, the setup won't be touched.

It is common practice to put shared values from the setup in the context which
can then be used later in the test and control cases.

## Testing

When you're testing you're application, it is important to see all the issues.
With the panic repression and random runs, this is impossible. For this reason,
there is the option `TestMode` available. Note, this should only be used whilst
testing!

The common case to set this is use the `Init()` function in your test helpers
to set this option.

```go
package application_test

import (
	"testing"

	"github.com/jelmersnoeck/experiment"
)

func init() {
	experiment.Init(experiment.TestMode())
}

// wherever you use an experiment, it will now panic on mismatch, panic when
// your code throws a panic and run all tests, regardless of other options.
```
