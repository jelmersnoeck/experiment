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

func beforeFilter(ctx context.Context) context.Context {
	return ctx
}
