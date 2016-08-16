package experiment

import "context"

type (
	behaviour struct {
		name string
		fnc  BehaviourFunc
	}

	// BehaviourFunc represents the type of function one can use as a
	// bahavioural function. We will inject a context which can contain multiple
	// key/value pairs and we expect it to return an interface and error. When
	// the function is linked to a control behaviour, the returned interface and
	// error will be used to be returned when the experiment is run.
	BehaviourFunc func(context.Context) (interface{}, error)
)

func newBehaviour(name string, fnc BehaviourFunc) *behaviour {
	return &behaviour{name, fnc}
}
