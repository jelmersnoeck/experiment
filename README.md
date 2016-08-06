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
var (
    myExperiment = newExperiment()
)

func main() {
	runner, err := myExperiment.Runner()
	if err != nil {
		fmt.Println(err)
		return
	}

    obs := runner.Run(nil)
	str = obs.Control().Value.(string)
	fmt.Println(str)
}

func newExperiment() *experiment.Experiment
	exp := experiment.New(
        experiment.DefaultConfig("my-test"),
    )

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

    return exp
)
```

First, we create an experiment with a new name. This will identify the
experiment later on in our publishers.

We then add a control function to the experiment. This is the functionality you
are currently using in your codebase and want to evaluate against. The `Control`
method is of the same structure as the `Test` method, in which it takes a
`context.Context` to pass along any data and expects an interface and error as
return value.

The next step is to define tests. These tests are to see if the newly refactored
code performs better and yields the same output. The sampling rate for these
tests can be configured as mentioned later on in the config section.

Once he setup is complete, it is time to run our experiment. To do so, we
request a runner from the experiment. We can then use some extra options
(discussed in the `Runner` section) and run the experiment.

The `Runner` will return an `Observations` type, which is a set of observations.
This includes our control observation - which can be accessed via `Control()` -
and our test observations.

Once the setup is complete, it is time to run our experiment. This will run our
control and tests(if applicable) in a random fashion. This means that one time
the control could be run first, another time the test case could be run first.
This is done so to avoid any accidental behavioural changes in any of the
control or test code.

## Limitations

### Stateless

Due to the fact that it is not guaranteed that a test will run every time or in
what order a test will run, it is suggested that experiments only do stateless
changes.

Tests also run concurrently next to each other, so it is important to keep this
in mind that your data should be concurrently accessible.

## Runner

After creating an experiment, we can request a runner from it. The runner is
resonsible for actually running the tests. Unlike an Experiment, a runner is
not safe for concurrent usage and should be created for each concurrent request.

### Disable

The `Disable(bool)` method is in place for checks when you might not want to run
an experiment at any cost. This overrules the `Force(bool)` method as well.

### Force

`Force(bool)` allows you to force run an experiment and overrules all other
options, apart from the `Disable(bool)` method.

### Run

`Run(context.Context)` will run the experiment and return a set of observations.

## Observation

An Observation contains several attributes. The first one is the `Value`. This
is the value which is returned by the control function that is specified. There
is also an `Error` attribute available, which contains the error returned.

## Errors

### Regular errors

When a `BehaviourFunc` returns an error, this error will be attached to the
`Observation` under the `Error` value.

### Panics

When the control panics, this panic will be respected and actually be triggered.
When a test function panics, the experiment will swallow this and add the panic
to the `Panic` attribute on the `Observation`.

## Config

When creating a new experiment, one can add several options. Some of them have
default values.

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

## Results

Tests results can be obtained by creating a new Results handler.

```go
func main () {
    // set up experiment and runner
    obs := runner.Run(nil)

    res := experiment.NewResult(obs, comparisonMethod)

    fmt.Println(res.Control()) // prints the control observation
    fmt.Println(res.Candidates()) // prints the observations that passes the comparison method check
    fmt.Println(res.Mismatches()) // prints the mismatches
}

func comparisonMethod(ctrl Observation, test Observation) bool {
    // do comparison logic here
}
```

It is important to note that this could be an expensive method to run. It is
advised to run this in a goroutine.

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
