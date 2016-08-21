package experiment_test

import (
	"testing"

	"github.com/jelmersnoeck/experiment"
)

func TestNewResult_NoComparison(t *testing.T) {
	res := experiment.NewResult(testObservations(), nil)

	if len, can := 2, len(res.Candidates()); len != can {
		t.Fatalf("Expected `%d` candidates, got `%d`", len, can)
	}

	if len(res.Mismatches()) != 0 {
		t.Fatalf("Expected `Mismatches()` to be empty")
	}
}

func TestNewResult_WithComparison(t *testing.T) {
	res := experiment.NewResult(testObservations(), comparisonMethod)

	if len, can := 1, len(res.Candidates()); len != can {
		t.Fatalf("Expected `%d` candidates, got `%d`", len, can)
	}

	if len, can := 1, len(res.Mismatches()); len != can {
		t.Fatalf("Expected `%d` mismatches, got `%d`", len, can)
	}
}

func testObservations() experiment.Observations {
	return experiment.Observations{
		"control": experiment.Observation{Value: "correct-test"},
		"test1":   experiment.Observation{Value: "incorrect"},
		"test2":   experiment.Observation{Value: "correct-test"},
	}
}

func dummyTestFunc(ctx experiment.Context) (interface{}, error) {
	return "test", nil
}

func dummyCompareTestFunc(ctx experiment.Context) (interface{}, error) {
	return "control", nil
}

func dummyControlFunc(ctx experiment.Context) (interface{}, error) {
	return "control", nil
}

func comparisonMethod(c experiment.Observation, t experiment.Observation) bool {
	return c.Value.(string) == t.Value.(string)
}
