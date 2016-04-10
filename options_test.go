package experiment

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/stretchr/testify/assert"
)

func TestDefaultOptions(t *testing.T) {
	defaults := newOptions()

	assert.Equal(t, "", defaults.name, "Default name")
	assert.Equal(t, float64(10), defaults.percentage, "Default percentage")
	assert.True(t, defaults.enabled, "Default enabler")
	assert.False(t, defaults.testMode, "Default testMode")
	assert.Nil(t, defaults.comparison, "Default comparison method")
	assert.NotNil(t, defaults.ctx, "Default context")
}

func TestOptions_Name(t *testing.T) {
	ops := newOptions(name("test-options-name"))
	assert.Equal(t, "test-options-name", ops.name, "Overwriting name")
}

func TestOptions_TestMode(t *testing.T) {
	ops := newOptions(TestMode())
	assert.True(t, ops.testMode)
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

func TestOptions_Context(t *testing.T) {
	val := "foo"
	ctx := context.WithValue(context.Background(), "test-ctx", val)

	ops := newOptions(Context(ctx))
	ctxVal := ops.ctx.Value("test-ctx")
	assert.Equal(t, val, ctxVal)
}
