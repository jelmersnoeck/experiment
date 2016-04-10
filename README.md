# Experiment

Experiment is a Go package to test and evaluate new code paths without
interfering with the users end result.

This is inspired by the [GitHub Scientist gem](https://github.com/github/scientist).

## Usage

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
