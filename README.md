# Experiment

Experiment is a Go package to test and evaluate new code paths without
interfering with the users end result.

This is inspired by the [GitHub Scientist gem](https://github.com/github/scientist).

## Usage

```go
func main() {
    exp, err := experiment.New(
        experiment.Name("my-test"),
        experiment.Enabled(shouldRunTest()),
        experiment.Percentage(10),
        experiment.Compare(comparisonMethod),
    )
    if err != nil {
        fmt.Println(err)
        return
    }

	exp.Control(func() interface{} {
		return "my-text"
	})
	exp.Test(func() interface{} {
		buf := bytes.NewBufferString("")
		buf.WriteString("new")
		buf.Write([]byte(`-`))
		buf.WriteString("text")

		return string(buf.Bytes())
	})

	exp.Run()
}

func shouldRunTest() bool {
	return os.Getenv("ENV") == "prod"
}

func comparisonMethod(control interface{}, test interface{}) bool {
	c := control.(string)
	t := test.(string)

	return c == t
}
```
