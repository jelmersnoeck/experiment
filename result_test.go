package experiment_test

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/jelmersnoeck/experiment"
	"github.com/stretchr/testify/assert"
)

func TestNewResult_NoComparison(t *testing.T) {
	exp := experiment.New("result-test")
	runExperiment(exp)

	res := experiment.NewResult(exp)
	assert.Len(t, res.Candidates(), 2)
	assert.NotNil(t, res.Control())
	assert.Empty(t, res.Mismatches())
}

func TestNewResult_WithComparison(t *testing.T) {
	exp := experiment.New(
		"result-test",
		experiment.Compare(comparisonMethod),
	)
	runExperiment(exp)

	res := experiment.NewResult(exp)
	assert.Len(t, res.Candidates(), 2)
	assert.NotNil(t, res.Control())
	assert.Len(t, res.Mismatches(), 1)
}

func runExperiment(exp *experiment.Experiment) {
	exp.Control(dummyControlFunc)
	exp.Test("test1", dummyTestFunc)
	exp.Test("test2", dummyCompareTestFunc)
	exp.Run()
}

func dummyTestFunc(ctx context.Context) (interface{}, error) {
	return "test", nil
}

func dummyCompareTestFunc(ctx context.Context) (interface{}, error) {
	return "control", nil
}

func dummyControlFunc(ctx context.Context) (interface{}, error) {
	return "control", nil
}

func comparisonMethod(c experiment.Observation, t experiment.Observation) bool {
	return c.Value().(string) == t.Value().(string)
}
