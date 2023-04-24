package experiment_test

import (
	"fmt"

	"github.com/jelmersnoeck/experiment"
)

func ExampleLogPublisher() {
	exp := experiment.New[string]().
		WithPublisher(experiment.NewLogPublisher[string]("publisher", &fmtLogger{}))

	exp.Control(func() (string, error) {
		return "Hello world!", nil
	})

	result, err := exp.Run()
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
