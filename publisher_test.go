package experiment_test

import (
	"context"
	"fmt"

	"github.com/jelmersnoeck/experiment/v3"
)

func ExampleLogPublisher() {
	exp := experiment.New[string]().
		WithPublisher(experiment.NewLogPublisher[string]("publisher", &fmtLogger{}))

	exp.Control(func(context.Context) (string, error) {
		return "Hello world!", nil
	})

	result, err := exp.Run(context.Background())
	if err != nil {
		panic(err)
	} else {
		fmt.Println(result)
	}

	// Output: Hello world!
}

type fmtLogger struct{}

func (l *fmtLogger) Printf(s string, a ...interface{}) {
	fmt.Printf(s, a...)
}
