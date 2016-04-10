package experiment

import (
	"errors"
	"testing"

	"golang.org/x/net/context"

	"github.com/stretchr/testify/assert"
)

func TestExperiment_New(t *testing.T) {
	exp := New("experiment-test")

	assert.Equal(t, "experiment-test", exp.Name(), "Experiment name from opts")
}

func TestExperiment_Control(t *testing.T) {
	exp := New("control-test")
	assert.Empty(t, exp.behaviours)

	err := exp.Control(dummyControlFunc)
	assert.NotEmpty(t, exp.behaviours)
	assert.Nil(t, err)

	err = exp.Control(dummyControlFunc)
	assert.NotNil(t, err)
	assert.Len(t, exp.behaviours, 1)
}

func TestExperiment_Test(t *testing.T) {
	exp := New("control-test")
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

func TestExperiment_Run_NoControl(t *testing.T) {
	exp := New("control-test")
	exp.Test("test-1", dummyTestFunc)

	_, err := exp.Run()
	assert.IsType(t, err, MissingControlError)
}

func TestExperiment_Run_NoTest(t *testing.T) {
	exp := New("control-test")
	exp.Control(dummyControlFunc)

	_, err := exp.Run()
	assert.IsType(t, err, MissingTestError)
}

func TestExperiment_Run(t *testing.T) {
	exp := New("control-test")

	exp.Control(dummyControlFunc)
	exp.Test("test-1", dummyTestFunc)

	obs, err := exp.Run()

	assert.Nil(t, err)
	assert.Equal(t, obs.Value().(string), "control")
}

func TestExperiment_Run_WithTestPanic(t *testing.T) {
	exp := New("control-test")

	exp.Control(dummyControlFunc)
	exp.Test("panic-test", dummyTestPanicFunc)

	obs, err := exp.Run()

	assert.Nil(t, err)
	assert.Equal(t, obs.Value().(string), "control")
	assert.Len(t, exp.observations, 2)

	panicObs := exp.observations["panic-test"]
	assert.NotNil(t, panicObs.Panic)
}

func dummyTestFunc(ctx context.Context) (interface{}, error) {
	return "test", nil
}

func dummyControlFunc(ctx context.Context) (interface{}, error) {
	return "control", nil
}

func dummyTestErrorFunc(ctx context.Context) (interface{}, error) {
	return "test", errors.New("error")
}

func dummyTestPanicFunc(ctx context.Context) (interface{}, error) {
	panic("test")
	return "panic", nil
}
