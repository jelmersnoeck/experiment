package experiment_test

import (
	"testing"

	"github.com/jelmersnoeck/experiment"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	df := experiment.DefaultConfig()
	require.EqualValues(t, 100, df.Percentage)
}
