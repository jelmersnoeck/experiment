package experiment

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/stretchr/testify/assert"
)

func TestExperiment_New_NoName(t *testing.T) {
	_, err := New()

	assert.NotNil(t, err, "Experiment without name error")
	assert.IsType(t, NoNameError, err, "Experiment without name error type")
}

func TestExperiment_New(t *testing.T) {
	exp, err := New(Name("experiment-test"))

	assert.Nil(t, err)
	assert.Equal(t, "experiment-test", exp.Name(), "Experiment name from opts")
}

func TestExperiment_Control(t *testing.T) {
	exp, _ := New(Name("control-test"))
	assert.Empty(t, exp.behaviours)

	err := exp.Control(dummyTestFunc)
	assert.NotEmpty(t, exp.behaviours)
	assert.Nil(t, err)

	err = exp.Control(dummyTestFunc)
	assert.NotNil(t, err)
	assert.Len(t, exp.behaviours, 1)
}

func TestExperiment_Test(t *testing.T) {
	exp, _ := New(Name("control-test"))
	assert.Empty(t, exp.behaviours)

	err := exp.Test("first", dummyTestFunc)
	assert.Nil(t, err)
	assert.Len(t, exp.behaviours, 1)

	err = exp.Test("first", dummyTestFunc)
	assert.NotNil(t, err)
	assert.Len(t, exp.behaviours, 1)

	err = exp.Test("second", dummyTestFunc)
	assert.Nil(t, err)
	assert.Len(t, exp.behaviours, 2)
}

func dummyTestFunc(ctx context.Context) (interface{}, error) {
	return nil, nil
}
