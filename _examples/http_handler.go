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
		exp := experiment.New[string](
			experiment.WithPercentage(50),
			experiment.WithConcurrency(),
		).WithPublisher(&experiment.LogPublisher[string]{})

		exp.Before(func() error {
			fmt.Println("before")
			return nil
		})

		exp.Control(func() (string, error) {
			return fmt.Sprintf("Hello %s", r.URL.Query().Get("name")), nil
		})

		exp.Candidate("foo", func() (string, error) {
			fmt.Println("foo")
			return "Hello foo", nil
		})

		exp.Candidate("bar", func() (string, error) {
			fmt.Println("bar")
			return "Hello bar", nil
		})

		exp.Candidate("baz", func() (string, error) {
			return "", errors.New("bar")
		})

		exp.Compare(func(control, candidate string) bool {
			fmt.Printf("Comparing '%s' with '%s'\n", control, candidate)
			return control == candidate
		})

		exp.Clean(func(c string) string {
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
			w.Write([]byte(result))
		}
	}
}
