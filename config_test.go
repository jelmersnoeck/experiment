package experiment_test

import (
	"testing"

	"github.com/jelmersnoeck/experiment"
)

func TestDefaultConfig(t *testing.T) {
	df := experiment.DefaultConfig()

	if val, pct := float32(100), df.Percentage; val != pct {
		t.Fatalf("Expected percentage to be `%f`, got `%f`", val, pct)
	}
}
