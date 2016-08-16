package experiment_test

import (
	"testing"

	"github.com/jelmersnoeck/experiment"
)

func TestObservation_Control(t *testing.T) {
	obs := experiment.Observations{
		"control": experiment.Observation{Name: "control"},
		"next":    experiment.Observation{Name: "next"},
	}

	if exp, val := "control", obs.Control().Name; exp != val {
		t.Fatalf("Expected control name to be `%s`, got `%s`", exp, val)
	}

}

func TestObservation_Tests(t *testing.T) {
	obs := experiment.Observations{
		"control": experiment.Observation{Name: "control"},
		"next":    experiment.Observation{Name: "next"},
	}

	if exp, len := 1, len(obs.Tests()); exp != len {
		t.Fatalf("Expectes tests length to be `%d`, got `%d`", exp, len)
	}
	if exp, val := "next", obs.Tests()[0].Name; exp != val {
		t.Fatalf("Expected test name to be `%s`, got `%s`", exp, val)
	}
}

func TestObservation_Find(t *testing.T) {
	obs := experiment.Observations{
		"control": experiment.Observation{Name: "control"},
		"next":    experiment.Observation{Name: "next"},
	}

	if exp, val := "next", obs.Find("next").Name; exp != val {
		t.Fatalf("Expected observation name to be `%s`, got `%s`", exp, val)
	}
	if exp, val := "control", obs.Find("control").Name; exp != val {
		t.Fatalf("Expected obserfvation name to be `%s`, got `%s`", exp, val)
	}
}
