package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jelmersnoeck/experiment"
)

type handleFunc func(http.ResponseWriter, *http.Request)

func main() {
	http.HandleFunc("/hello", exampleHandler())
	log.Fatal(http.ListenAndServe(":12345", nil))
}

func exampleHandler() handleFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		exp := experiment.New(
			experiment.WithName("example-test"),
			experiment.WithPercentage(50),
			experiment.WithPublisher(nil),
			experiment.WithConcurrency(),
		)

		exp.Before(func() error {
			return nil
		})

		exp.Control(func() (interface{}, error) {
			return fmt.Sprintf("Hello %s", r.URL.Query().Get("name")), nil
		})

		exp.Candidate("foo", func() (interface{}, error) {
			return "Hello foo", nil
		})

		exp.Candidate("bar", func() (interface{}, error) {
			return "Hello bar", nil
		})

		exp.Candidate("baz", func() (interface{}, error) {
			return "Hello baz", nil
		})

		exp.Compare(func(control interface{}, candidate interface{}) bool {
			return control.(string) == candidate.(string)
		})

		exp.Clean(func(c interface{}) {
			// do cleanup
		})

		// exp.Force(user.IsAdmin())

		// exp.Ignore(env.Test())

		result, err := exp.Run()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			w.Write([]byte(result.(string)))
		}
	}
}
