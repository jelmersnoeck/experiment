package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jelmersnoeck/experiment/v3"
)

type handleFunc func(http.ResponseWriter, *http.Request)

func main() {
	log.Print("Starting server, access http://0.0.0.0:12345/hello")
	http.HandleFunc("/hello", exampleHandler())
	log.Fatal(http.ListenAndServe(":12345", nil))
}

func exampleHandler() handleFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exp := experiment.New[string](
			experiment.WithPercentage(50),
			experiment.WithConcurrency(),
			experiment.WithTimeout(500*time.Millisecond),
		).WithPublisher(&experiment.LogPublisher[string]{})

		exp.Before(func(context.Context) error {
			fmt.Println("before")
			return nil
		})

		exp.Control(func(context.Context) (string, error) {
			return fmt.Sprintf("Hello %s", r.URL.Query().Get("name")), nil
		})

		exp.Candidate("foo", func(context.Context) (string, error) {
			fmt.Println("foo")
			return "Hello foo", nil
		})

		exp.Candidate("bar", func(context.Context) (string, error) {
			fmt.Println("bar")
			return "Hello bar", nil
		})

		exp.Candidate("baz", func(context.Context) (string, error) {
			return "", errors.New("bar")
		})

		exp.Candidate("timeout", func(ctx context.Context) (string, error) {
			select {
			case <-time.Tick(time.Second):
				return "Waited a full second!", nil
			case <-ctx.Done():
				return "Timeout hit", ctx.Err()
			}
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

		ctx := context.Background()
		result, err := exp.Run(ctx)
		_ = exp.Publish(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			w.Write([]byte(result))
		}
	}
}
