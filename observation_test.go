package experiment_test

import (
	"testing"

	"github.com/jelmersnoeck/experiment"
	"github.com/stretchr/testify/require"
)

func TestObservation_Control(t *testing.T) {
	obs := experiment.Observations{
		"control": experiment.Observation{Name: "control"},
		"next":    experiment.Observation{Name: "next"},
	}

	require.Equal(t, "control", obs.Control().Name)
}

func TestObservation_Candidates(t *testing.T) {
	obs := experiment.Observations{
		"control": experiment.Observation{Name: "control"},
		"next":    experiment.Observation{Name: "next"},
	}

	require.Len(t, obs.Candidates(), 1)
	require.Equal(t, "next", obs.Candidates()[0].Name)
}

func TestObservation_Find(t *testing.T) {
	obs := experiment.Observations{
		"control": experiment.Observation{Name: "control"},
		"next":    experiment.Observation{Name: "next"},
	}

	require.Equal(t, "next", obs.Find("next").Name)
	require.Equal(t, "control", obs.Find("control").Name)
}
