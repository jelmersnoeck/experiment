package experiment

import (
	"errors"
	"testing"

	"golang.org/x/net/context"

	"github.com/stretchr/testify/require"
)

func TestExperiment_Control(t *testing.T) {
	exp := newExperiment(DefaultConfig("test"))
	require.Empty(t, exp.behaviours)

	err := exp.Control(dummyControlFunc)
	require.NotEmpty(t, exp.behaviours)
	require.Nil(t, err)

	err = exp.Control(dummyControlFunc)
	require.NotNil(t, err)
	require.Len(t, exp.behaviours, 1)
}

func TestExperiment_Test(t *testing.T) {
	exp := newExperiment(DefaultConfig("test"))
	require.Empty(t, exp.behaviours)

	err := exp.Test("first", dummyTestFunc)
	require.Nil(t, err)
	require.Len(t, exp.behaviours, 1)

	err = exp.Test("first", dummyTestFunc)
	require.NotNil(t, err)
	require.Len(t, exp.behaviours, 1)

	err = exp.Test("second", dummyTestFunc)
	require.Nil(t, err)
	require.Len(t, exp.behaviours, 2)
}

func TestExperiment_Run_NoControl(t *testing.T) {
	exp := newExperiment(DefaultConfig("test"))
	exp.Test("test-1", dummyTestFunc)

	_, err := exp.Run(nil)
	require.IsType(t, err, ErrMissingControl)
}

func TestExperiment_Run(t *testing.T) {
	exp := newExperiment(DefaultConfig("test"))

	exp.Control(dummyControlFunc)
	exp.Test("test-1", dummyTestFunc)

	obs, err := exp.Run(nil)

	require.Nil(t, err)
	require.NotNil(t, obs)
	require.Equal(t, obs.Control().Value.(string), "control")
}

func TestExperiment_Run_WithTestPanic(t *testing.T) {
	exp := newExperiment(DefaultConfig("test"))

	exp.Control(dummyControlFunc)
	exp.Test("panic-test", dummyTestPanicFunc)

	obs, err := exp.Run(nil)

	require.Nil(t, err)
	require.Equal(t, obs.Control().Value.(string), "control")
	require.Len(t, obs, 2)

	panicObs := obs.Find("panic-test")
	require.NotNil(t, panicObs.Panic)
}

func TestExperiment_Run_WithContext(t *testing.T) {
	val := "my-context-test"
	ctx := context.WithValue(context.Background(), "ctx-test", val)

	exp := newExperiment(DefaultConfig("test"))
	exp.Control(dummyContextTestFunc)

	obs, err := exp.Run(ctx)

	require.Nil(t, err)
	require.Equal(t, obs.Control().Value.(string), val)
}

func TestExperiment_Run_Before(t *testing.T) {
	beforeFunc := func(ctx context.Context) context.Context {
		return context.WithValue(ctx, "my-key", "my-value")
	}
	checkFunc := func(ctx context.Context) (interface{}, error) {
		str := ctx.Value("my-key")

		require.Equal(t, "my-value", str)
		return nil, nil
	}

	cfg := DefaultConfig("test")
	cfg.AddBeforeFilter(beforeFunc)

	exp := newExperiment(cfg)
	exp.Control(checkFunc)

	exp.Run(context.Background())
}

func BenchmarkExperiment_Run(b *testing.B) {
	exp := newExperiment(DefaultConfig("benchmark-test"))

	exp.Control(dummyControlFunc)
	exp.Test("first", dummyTestFunc)
	exp.Test("second", dummyTestFunc)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			exp.Run(nil)
		}
	})
}

func newExperiment(cfg *Config) *Experiment {
	return &Experiment{
		Config: cfg,
	}
}

func dummyContextTestFunc(ctx context.Context) (interface{}, error) {
	return ctx.Value("ctx-test"), nil
}

func dummyTestFunc(ctx context.Context) (interface{}, error) {
	return "test", nil
}

func dummyControlFunc(ctx context.Context) (interface{}, error) {
	return "control", nil
}

func dummyTestErrorFunc(ctx context.Context) (interface{}, error) {
	return "test", errors.New("error")
}

func dummyTestPanicFunc(ctx context.Context) (interface{}, error) {
	panic("test")
	return "panic", nil
}
