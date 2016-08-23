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

    obs := runner.Run(context.TODO())
	str = obs.Control().Value.(string)
	fmt.Println(str)
}

func newExperiment() *experiment.Experiment
	exp := experiment.New(
        experiment.DefaultConfig(),
    )

	exp.Control(func(ctx experiment.Context) (interface{}, error) {
		return "my-text", nil
	})
	exp.Test("buffer", func(ctx experiment.Context) (interface{}, error) {
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

## Context

[Context](https://godoc.org/context) has been added to Go in version 1.7. For backwards compatibility,
the Context interface has been copied to this library instead of relying on the
Go 1.7 internal interface or the older `/x/net/context` interface.

Note: when using a Context, one should always use an actual implementation of a
context and not use `nil`. If the implementation isn't decided yet, use
`context.TODO()`.

## Limitations and caveats

### Stateless

Due to the fact that it is not guaranteed that a test will run every time or in
what order a test will run, it is suggested that experiments only do stateless
changes.

Tests also run concurrently next to each other, so it is important to keep this
in mind that your data should be concurrently accessible.

### Concurrent access to context

When annotating the context which is passed to the runner, one should think
about the implications of what changing this data would mean. A general rule is
to only *read* data from `context.Context` and not change it.

Below is an example that demonstrates missbehaviour:

```go
type Test struct {
    Value string
}

func main() {
    // setup
    exp.Control(controlFunc)
    exp.Test("test", testFunc)

    run, _ := exp.Runner()

    value := &Test{"Foo"}
    ctx := context.WithValue(context.Background(), "key", value)
    runner.Run(value)
}

func controlFunc(ctx experiment.Context) (interface{}, error) {
    test := context.Value(ctx, "key").(*Test)

    // The below value could either be "Foo" or "Bar" due to changing the `key`
    // value, which is a pointer (the context points to the same reference in
    // both the controlFunc and testFunc).
    fmt.Println(test.Value)

    return test, nil
}

func testFunc(ctx experiment.Context) (interface{}, error) {
    test := context.Value(ctx, "key").(*Test)
    test.Value = "Bar"

    return test, nil
}
```

### Performance

Although all the experiments run at the same time (with goroutines), it could be that new tests introduce a performance degradation. New tests should be rolled out slowly and monitored closely. Using the `Config` `Percentage` option is a good first step for this.

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
        Percentage: 25,
    }
	exp := experiment.New(cfg)

	// add control/tests

	obs, err := exp.Run(context.TODO())
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
        Percentage: 25,
        BeforeFilters: []BeforeFilter{mySetup},
    }
	exp := experiment.New(cfg)

	// do more experiment setup and run it
}

func mySetup(ctx experiment.Context) experiment.Context {
	expensive := myExpensiveSetup()
	return context.WithValue(ctx, "my-thing", expensive)
}

func myControlFunc(ctx experiment.Context) (interface{}, error) {
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
    obs := runner.Run(context.TODO())

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

## Publisher

The publisher is what makes the experiment useful. It allows you to see how an
experiment performs and what the impact is.

The publisher depends on 3 methods:

- `Increment(string)` for incrementing error and panic counts
- `Count(string, interface{})` for counting candidates and mismatches
- `Timing(string, interface{})` for measuring run durations

### Keys

Error and panic counts are captured with the keys `<test-name>.panics.incr` and
`<test-name>.errors.incr`.

Candidates and mismatches are captured with the keys `candidates.count` and
`mismatches.count`.

Timings are captured with the key `<test-name>.time`.


*Note:* The control function has the `<test-name>` `control`.

### Concurrent access

The publisher is safe for concurrent usage and should be used in such way. This
means that one should create only one publisher (for example with a statsd
client) and reuse this publisher across all requests.

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
    experiment.TestMode = true
}

// wherever you use an experiment, it will now panic on mismatch, panic when
// your code throws a panic and run all tests, regardless of other options.
```
