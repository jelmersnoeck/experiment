package experiment_test

import (
	"fmt"

	"github.com/jelmersnoeck/experiment"
)

func ExampleLogPublisher() {
	exp := experiment.New(
		experiment.WithPublisher(experiment.NewLogPublisher("publisher", &fmtLogger{})),
	)

	exp.Control(func() (interface{}, error) {
		return "Hello world!", nil
	})

	result, err := exp.Run()
	if err != nil {
		panic(err)
	} else {
		fmt.Println(result.(string))
	}

	// Output: Hello world!
}

type fmtLogger struct{}

func (l *fmtLogger) Printf(s string, a ...interface{}) {
	fmt.Printf(s, a...)
}
