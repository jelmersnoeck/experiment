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
