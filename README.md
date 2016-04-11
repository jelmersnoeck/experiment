# Experiment

Experiment is a Go package to test and evaluate new code paths without
interfering with the users end result.

This is inspired by the [GitHub Scientist gem](https://github.com/github/scientist).

## Usage

Below is the most basic example on how to create an experiment. We'll see if
using a string with the notation of `""` compares to using a byte buffer and
we'll evaluate the time difference.

```go
func main() {
	exp := experiment.New("my-test")

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

	obs, err := exp.Run()
	if err != nil {
		fmt.Println(err)
		return
	}
	str = obs.Value().(string)
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

The run method also returns an observation, which contains the return value and
error from the control method. If something went wrong trying to run the
experiment, this will also return the error it encountered. The control code
is always executed as is in comparison to the test code. If an error happens
within the test code which causes an exception, this will be swallowed and added
to the observation.

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

The `Run()` method for an experiment executes the experiment. This means that it
will run the control and potentially the tests.

The control will be run no matter what. The tests might run depending on several
options (see listed below).

The `Run()` method will return an Observation and an error. When an error is
returned, it means that the control couldn't be run for some reason.

The Observation contains several methods. The first one is the `Value()`. This
is the value which is returned by the control function that is specified. There
is also an `Error()` method available, which contains the error returned.

## Options

When creating a new experiment, one can add several options. Some of them have
default values.

- Comparison (nil)
- Percentage (10)
- Enabled (true)

### Comparison

By default, the duration and return values will be captured. If you want to
conclude the experiment by comparing results, a `Compare` option can be given
when creating a new experiment. Mismatches will then be recorded in the `Result`
which can be used to be published.

If no `Compare` option is given, no mismatches will be recorded, only durations.

```go
func main() {
	exp := experiment.New(
		"my-experiment",
		experiment.Compare(comparisonMethod),
	)

	// add control/tests

	exp.Run()

	result := exp.Result()
	fmt.Println(result.Mismatches)
}

func comparisonMethod(control experiment.Observation, test experiment.Observation) bool {
	c := control.Value().(string)
	t := test.Value().(string)

	return c == t
}
```

### Percentage

Sometimes, you don't want to run the experiment for every request. To do this,
one can set the percentage rate of which we'll run our test cases.

The control will always be run when using the `Run()` method. This option just
disables the test cases to be run.


```go
func main() {
	exp := experiment.New(
		"my-experiment",
		experiment.Percentage(10),
	)

	// add control/tests

	obs, err := exp.Run()
	if err != nil {
		fmt.Println(err)
		return
	}

	str = obs.Value().(string)
	fmt.Println(str)
}
```

### Enabled

While refactoring code, it might be possible that a certain code path is not
available yet. To accomplish this, there is an `Enabled(bool)` option available.

```go
func main() {
	// do set up

	u := User.Get(5)

	exp := experiment.New(
		"my-test",
		experiment.Enabled(shouldRunExperiment(u)),
	)

	// run the experiment
}

func shouldRunExperiment(user User) bool {
	return user.IsConfirmed()
}
```

In this case, if the user is not confirmed yet, we will not run the experiment.

### Context

When using a context for your request, you might have information that you need
within your test. Using the `Context()` option, you can now set a context that
will be used to pass along to your test functions.

```go
func main() {
	ctx := context.WithValue(context.Background(), "key", "value")

	exp := experiment.New(
		"context-example",
		experiment.Context(ctx),
	)
	exp.Control(myControlFunc)

	// do more experiment setup and run it
}

func myControlFunc(ctx context.Context) (interface{}, error) {
	key := ctx.Value("key")

	return key, nil
}
```

### Before

When an expensive setup is required to do the test, we don't always want to run
the setup until we actually execute the test case. The `Before` option allows us
to set this up. This means that when an experiment is not run, this set up will
not be executed.

To do this, we make use of context.

```go
func main() {
	exp := experiment.New(
		"context-example",
		experiment.Before(mySetup),
	)

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

### Publisher

Once the experiment has run, it is useful to see the results. To do so, there
is a `ResultPublisher` interface available. This has one method,
`Publish(Result)` which will take care of publishing the result to the chosen
output.

Multiple publishers can be configured for a single experiment. For example, one
could use a statsd publisher to pubish duration metrics to statsd and a Redis
publisher to store the differences between the control and test results.

```go
func main() {
	exp := experiment.New(
		"context-example",
		experiment.Publisher(myPublisher{}),
		experiment.Publisher(redisPublisher{}),
	)

	// more experiment setup and run

	// this will publish the results to `myPublisher` and `redisPublisher`.
	exp.Publish()
}
```

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
