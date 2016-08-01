package experiment_test

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/jelmersnoeck/experiment"
	"github.com/stretchr/testify/require"
)

func TestNewResult_NoComparison(t *testing.T) {
	res := experiment.NewResult(testObservations(), nil)

	require.Len(t, res.Candidates(), 2)
	require.NotNil(t, res.Control())
	require.Empty(t, res.Mismatches())
}

func TestNewResult_WithComparison(t *testing.T) {
	res := experiment.NewResult(testObservations(), comparisonMethod)

	require.Len(t, res.Candidates(), 1)
	require.NotNil(t, res.Control())
	require.Len(t, res.Mismatches(), 1)
}

func testObservations() experiment.Observations {
	return experiment.Observations{
		"control": experiment.Observation{Value: "correct-test"},
		"test1":   experiment.Observation{Value: "incorrect"},
		"test2":   experiment.Observation{Value: "correct-test"},
	}
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
	return c.Value.(string) == t.Value.(string)
}
