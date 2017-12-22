package main

import (
	"errors"
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
			experiment.WithPercentage(50),
			experiment.WithPublisher(&experiment.LogPublisher{}),
			experiment.WithConcurrency(),
		)

		exp.Before(func() error {
			fmt.Println("before")
			return nil
		})

		exp.Control(func() (interface{}, error) {
			return fmt.Sprintf("Hello %s", r.URL.Query().Get("name")), nil
		})

		exp.Candidate("foo", func() (interface{}, error) {
			fmt.Println("foo")
			return "Hello foo", nil
		})

		exp.Candidate("bar", func() (interface{}, error) {
			fmt.Println("bar")
			return "Hello bar", nil
		})

		exp.Candidate("baz", func() (interface{}, error) {
			return nil, errors.New("bar")
		})

		exp.Compare(func(control interface{}, candidate interface{}) bool {
			fmt.Printf("Comparing '%s' with '%s'\n", control.(string), candidate.(string))
			return control.(string) == candidate.(string)
		})

		exp.Clean(func(c interface{}) interface{} {
			fmt.Println("cleanup")
			return c
		})

		exp.Force(r.URL.Query().Get("force") == "true")
		exp.Ignore(r.URL.Query().Get("ignore") == "true")

		result, err := exp.Run()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			w.Write([]byte(result.(string)))
		}
	}
}
