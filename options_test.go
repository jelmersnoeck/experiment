package experiment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultOptions(t *testing.T) {
	defaults := newOptions()

	assert.Equal(t, "", defaults.name, "Default name")
	assert.Equal(t, float64(10), defaults.percentage, "Default percentage")
	assert.True(t, defaults.enabled, "Default enabler")
	assert.Nil(t, defaults.comparison, "Default comparison method")
}

func TestOptions_Name(t *testing.T) {
	ops := newOptions(Name("test-options-name"))
	assert.Equal(t, "test-options-name", ops.name, "Overwriting name")
}

func TestOptions_Percentage(t *testing.T) {
	ops := newOptions(Percentage(5))
	assert.Equal(t, float64(5), ops.percentage, "Overwriting percentage")
}

func TestOptions_Enabled(t *testing.T) {
	ops := newOptions(Enabled(false))
	assert.False(t, ops.enabled, "Overwriting enabler")
}

func TestOptions_Compare(t *testing.T) {
	cmp := func(c Observation, t Observation) bool {
		return false
	}
	ops := newOptions(Compare(cmp))
	assert.NotNil(t, ops.comparison, "Overwriting comparison method")
}
