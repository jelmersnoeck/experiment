package experiment_test

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/jelmersnoeck/experiment"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	df := experiment.DefaultConfig("test")

	require.Equal(t, "test", df.Name)
	require.EqualValues(t, 100, df.Percentage)
}

func TestConfig_AddBeforeFilter(t *testing.T) {
	df := experiment.DefaultConfig("test")
	require.Len(t, df.BeforeFilters, 0)

	df.AddBeforeFilter(beforeFilter)
	require.Len(t, df.BeforeFilters, 1)
}

func TestConfig_SetComparisonMethod(t *testing.T) {
	df := experiment.DefaultConfig("test")
	require.Nil(t, df.Comparison)

	df.SetComparisonMethod(comparisonMethod)
	require.NotNil(t, df.Comparison)
}

func beforeFilter(ctx context.Context) context.Context {
	return ctx
}
