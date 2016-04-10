package experiment_test

import (
	"testing"

	"github.com/jelmersnoeck/experiment"
	"github.com/stretchr/testify/assert"
)

func TestExperiment_New_NoName(t *testing.T) {
	_, err := experiment.New()

	assert.NotNil(t, err, "Experiment without name error")
	assert.IsType(t, experiment.NoNameError, err, "Experiment without name error type")
}

func TestExperiment_New(t *testing.T) {
	exp, err := experiment.New(experiment.Name("experiment-test"))

	assert.Nil(t, err)
	assert.Equal(t, "experiment-test", exp.Name(), "Experiment name from opts")
}
