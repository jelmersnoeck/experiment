package experiment

import (
	"errors"
	"testing"

	"golang.org/x/net/context"
)

func TestExperiment_Control(t *testing.T) {
	exp := newExperiment(DefaultConfig())
	if len(exp.behaviours) != 0 {
		t.Fatalf("Expected behaviours to be empty, got `%d`", len(exp.behaviours))
	}

	err := exp.Control(dummyControlFunc)
	if err != nil {
		t.Fatalf("Expected error to be nil, got `%s`", err)
	}
	if len(exp.behaviours) == 0 {
		t.Fatalf("Expected behaviours not to be empty")
	}

	err = exp.Control(dummyControlFunc)
	if err == nil {
		t.Fatalf("Expected error not to be nil")
	}
	if exp, len := 1, len(exp.behaviours); exp != len {
		t.Fatalf("Expected `%d` behaviours, got `%d`", exp, len)
	}
}

func TestExperiment_Test(t *testing.T) {
	exp := newExperiment(DefaultConfig())
	if len(exp.behaviours) != 0 {
		t.Fatalf("Expected behaviours to be empty, got `%d`", len(exp.behaviours))
	}

	err := exp.Test("first", dummyTestFunc)
	if err != nil {
		t.Fatalf("Expected error to be nil, got `%s`", err)
	}
	if exp, len := 1, len(exp.behaviours); exp != len {
		t.Fatalf("Expected `%d` behaviours, got `%d`", exp, len)
	}

	err = exp.Test("first", dummyTestFunc)
	if err == nil {
		t.Fatalf("Expected error not to be nil")
	}
	if exp, len := 1, len(exp.behaviours); exp != len {
		t.Fatalf("Expected `%d` behaviours, got `%d`", exp, len)
	}

	err = exp.Test("second", dummyTestFunc)
	if err != nil {
		t.Fatalf("Expected error to be nil, got `%s`", err)
	}
	if exp, len := 2, len(exp.behaviours); exp != len {
		t.Fatalf("Expected `%d` behaviours, got `%d`", exp, len)
	}
}

func TestExperiment_Run_NoControl(t *testing.T) {
	exp := newExperiment(DefaultConfig())
	exp.Test("test-1", dummyTestFunc)

	_, err := exp.Runner()
	if err != ErrMissingControl {
		t.Fatalf("Expected error to be of type `ErrMissingControl`, got %T", err)
	}
}

func TestExperiment_Run(t *testing.T) {
	exp := newExperiment(DefaultConfig())

	exp.Control(dummyControlFunc)
	exp.Test("test-1", dummyTestFunc)

	runner, err := exp.Runner()
	if err != nil {
		t.Fatalf("Expected error to be nil, got `%s`", err)
	}

	obs := runner.Run(nil)
	if obs == nil {
		t.Fatalf("Expected observation not to be nil")
	}
	if exp, val := "control", obs.Control().Value.(string); exp != val {
		t.Fatalf("Expected control value to equal `%s`, got `%s`", exp, val)
	}
}

func TestExperiment_Runner_ControlFailure(t *testing.T) {
	exp := newExperiment(DefaultConfig())
	exp.Control(dummyTestErrorFunc)

	runner, err := exp.Runner()
	if err != nil {
		t.Fatalf("Expected error to be nil, got `%s`", err)
	}

	obs := runner.Run(nil)
	if obs.Control().Error == nil {
		t.Fatalf("Expected control error not to be nil")
	}
}

func TestExperiment_Run_WithTestPanic(t *testing.T) {
	exp := newExperiment(DefaultConfig())

	exp.Control(dummyControlFunc)
	exp.Test("panic-test", dummyTestPanicFunc)

	runner, err := exp.Runner()
	if err != nil {
		t.Fatalf("Expected error to be nil, got `%s`", err)
	}

	runner.Force(true)
	obs := runner.Run(nil)
	if exp, val := "control", obs.Control().Value.(string); exp != val {
		t.Fatalf("Expected control value to equal `%s`, got `%s`", exp, val)
	}
	if exp, val := 2, len(obs); exp != val {
		t.Fatalf("Expected `%d` observations, got `%d`", exp, val)
	}

	panicObs := obs.Find("panic-test")
	if panicObs.Panic == nil {
		t.Fatalf("Expected Panic not to be nil")
	}
}

func TestExperiment_Run_WithContext(t *testing.T) {
	val := "my-context-test"
	ctx := context.WithValue(context.Background(), "ctx-test", val)

	exp := newExperiment(DefaultConfig())
	exp.Control(dummyContextTestFunc)

	runner, err := exp.Runner()
	if err != nil {
		t.Fatalf("Expected error to be nil, got `%s`", err)
	}

	obs := runner.Run(ctx)
	if exp, val := "my-context-test", obs.Control().Value.(string); exp != val {
		t.Fatalf("Expected control value to equal `%s`, got `%s`", exp, val)
	}
}

func TestExperiment_Run_Before(t *testing.T) {
	beforeFunc := func(ctx Context) Context {
		return context.WithValue(ctx, "my-key", "my-value")
	}
	checkFunc := func(ctx Context) (interface{}, error) {
		if exp, val := "my-value", ctx.Value("my-key").(string); exp != val {
			t.Fatalf("Expected context string to be `%s`, got `%s`", exp, val)
		}
		return nil, nil
	}

	cfg := Config{
		Percentage:    100,
		BeforeFilters: []BeforeFilter{beforeFunc},
	}

	exp := newExperiment(cfg)
	exp.Control(checkFunc)

	runner, err := exp.Runner()
	if err != nil {
		t.Fatalf("Expected error to be nil, got `%s`", err)
	}

	runner.Run(nil)
}

func TestExperiment_Run_Percentage(t *testing.T) {
	cfg := Config{
		Percentage: 50,
	}

	exp := newExperiment(cfg)
	exp.Control(dummyControlFunc)
	exp.Test("first", dummyTestFunc)

	runner, err := exp.Runner()
	if err != nil {
		t.Fatalf("Expected error to be nil, got `%s`", err)
	}
	obs := runner.Run(nil)
	if exp, len := 2, len(obs); exp != len {
		t.Fatalf("Expected `%d` observations, got `%d`", exp, len)
	}

	runner, err = exp.Runner()
	if err != nil {
		t.Fatalf("Expected error to be nil, got `%s`", err)
	}
	obs = runner.Run(nil)
	if exp, len := 1, len(obs); exp != len {
		t.Fatalf("Expected `%d` observations, got `%d`", exp, len)
	}

	runner, err = exp.Runner()
	if err != nil {
		t.Fatalf("Expected error to be nil, got `%s`", err)
	}
	obs = runner.Run(nil)
	if exp, len := 2, len(obs); exp != len {
		t.Fatalf("Expected `%d` observations, got `%d`", exp, len)
	}

	runner, err = exp.Runner()
	if err != nil {
		t.Fatalf("Expected error to be nil, got `%s`", err)
	}
	obs = runner.Run(nil)
	if exp, len := 1, len(obs); exp != len {
		t.Fatalf("Expected `%d` observations, got `%d`", exp, len)
	}
}

func BenchmarkExperiment_Run(b *testing.B) {
	exp := newExperiment(DefaultConfig())

	exp.Control(dummyControlFunc)
	exp.Test("first", dummyTestFunc)
	exp.Test("second", dummyTestFunc)
	exp.Test("third", dummyTestFunc)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			runner, _ := exp.Runner()
			runner.Force(true)
			runner.Run(nil)
		}
	})
}

func newExperiment(cfg Config) *Experiment {
	return &Experiment{
		Config: cfg,
	}
}

func dummyContextTestFunc(ctx Context) (interface{}, error) {
	return ctx.Value("ctx-test"), nil
}

func dummyTestFunc(ctx Context) (interface{}, error) {
	return "test", nil
}

func dummyControlFunc(ctx Context) (interface{}, error) {
	return "control", nil
}

func dummyTestErrorFunc(ctx Context) (interface{}, error) {
	return "test", errors.New("error")
}

func dummyTestPanicFunc(ctx Context) (interface{}, error) {
	panic("test")
	return "panic", nil
}
