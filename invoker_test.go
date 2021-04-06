package invoker_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/kixelated/invoker"
	"github.com/stretchr/testify/require"
)

// Test with no tasks.
func TestRunEmpty(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	err := invoker.Run(ctx)
	require.Error(err)
}

// Test with all successes.
func TestRunSuccess(t *testing.T) {
	require := require.New(t)

	count := uint64(0)
	f := func(ctx context.Context) (err error) {
		atomic.AddUint64(&count, 1)
		return nil
	}

	err := invoker.Run(context.Background(), f, f, f)
	require.NoError(err)
	require.Equal(uint64(3), atomic.LoadUint64(&count))
}

// Test with an explicit cancel.
func TestRunCancel(t *testing.T) {
	require := require.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	count := uint64(0)
	f := func(ctx context.Context) (err error) {
		cancel() // cancel inside of the function to ensure invoker is started
		<-ctx.Done()
		atomic.AddUint64(&count, 1)
		return ctx.Err()
	}

	err := invoker.Run(ctx, f, f, f)
	require.Equal(context.Canceled, err)
	require.Equal(uint64(3), atomic.LoadUint64(&count))
}

// Test with an error result.
func TestRunError(t *testing.T) {
	require := require.New(t)

	errSample := fmt.Errorf("hello")

	count := uint64(0)
	f := func(ctx context.Context) (err error) {
		old := atomic.AddUint64(&count, 1)
		if old == 1 {
			return errSample
		}

		<-ctx.Done()
		return ctx.Err()
	}

	err := invoker.Run(context.Background(), f, f, f)
	require.Equal(errSample, err)
	require.Equal(uint64(3), atomic.LoadUint64(&count))
}

// Test with an asynchronous Add call.
func TestRunAdd(t *testing.T) {
	require := require.New(t)

	tasks := invoker.New()

	count := uint64(0)
	errSample := fmt.Errorf("hello")

	f2 := func(ctx context.Context) (err error) {
		atomic.AddUint64(&count, 1)
		return errSample
	}

	f := func(ctx context.Context) (err error) {
		atomic.AddUint64(&count, 1)

		// Each invocation adds another invocation that will error (and cancel).
		tasks.Add(f2)

		<-ctx.Done()
		return ctx.Err()
	}

	tasks.Add(f, f, f)

	err := tasks.Run(context.Background())
	require.Equal(errSample, err)
	require.Equal(uint64(6), atomic.LoadUint64(&count))
}

// Test with Add calls before running.
func TestRunAddPending(t *testing.T) {
	require := require.New(t)

	count := uint64(0)
	f := func(ctx context.Context) (err error) {
		atomic.AddUint64(&count, 1)
		return nil
	}

	tasks := invoker.New()
	tasks.Add(f, f, f)
	tasks.Add(f, f)

	err := tasks.Run(context.Background())
	require.NoError(err)
	require.Equal(uint64(5), atomic.LoadUint64(&count))
}

// Test reusing the invoker object.
func TestRunReuse(t *testing.T) {
	require := require.New(t)

	count := uint64(0)
	f := func(ctx context.Context) (err error) {
		atomic.AddUint64(&count, 1)
		return nil
	}

	tasks := invoker.New(f, f, f)

	err := tasks.Run(context.Background())
	require.NoError(err)
	require.Equal(uint64(3), atomic.LoadUint64(&count))

	err = tasks.Run(context.Background())
	require.Error(err)
}

// Test running the same invoker object twice.
func TestRunRunning(t *testing.T) {
	require := require.New(t)

	f := func(ctx context.Context) (err error) {
		<-ctx.Done()
		return ctx.Err()
	}

	tasks := invoker.New(f, f, f)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errs := make(chan error, 2)

	go func() {
		errs <- tasks.Run(ctx)
		cancel()
	}()

	errs <- tasks.Run(ctx)
	cancel()

	require.Equal(invoker.ErrRunning, <-errs)
	require.Equal(context.Canceled, <-errs)
}

// Test with not tasks.
func TestRaceEmpty(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	err := invoker.Race(ctx)
	require.Error(err)
}

// Test with tasks that succeed.
func TestRaceSuccess(t *testing.T) {
	require := require.New(t)

	count := uint64(0)
	f := func(ctx context.Context) (err error) {
		atomic.AddUint64(&count, 1)
		return nil
	}

	err := invoker.Race(context.Background(), f, f, f)
	require.NoError(err)
	require.Equal(uint64(3), atomic.LoadUint64(&count))
}

// Test with tasks that cancel.
func TestRaceCancel(t *testing.T) {
	require := require.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	count := uint64(0)
	f := func(ctx context.Context) (err error) {
		cancel() // cancel inside of the function to ensure invoker is started
		<-ctx.Done()
		atomic.AddUint64(&count, 1)
		return ctx.Err()
	}

	err := invoker.Race(ctx, f, f, f)
	require.Equal(context.Canceled, err)
	require.Equal(uint64(3), atomic.LoadUint64(&count))
}

// Test with tasks that (sometimes) error.
func TestRaceError(t *testing.T) {
	require := require.New(t)

	errSample := fmt.Errorf("hello")

	count := uint64(0)
	f := func(ctx context.Context) (err error) {
		old := atomic.AddUint64(&count, 1)
		if old == 1 {
			return errSample
		}

		<-ctx.Done()
		return ctx.Err()
	}

	err := invoker.Race(context.Background(), f, f, f)
	require.Equal(errSample, err)
	require.Equal(uint64(3), atomic.LoadUint64(&count))
}

// Test that the first result will cancel.
func TestRaceFirst(t *testing.T) {
	require := require.New(t)

	count := uint64(0)
	f := func(ctx context.Context) (err error) {
		old := atomic.AddUint64(&count, 1)
		if old == 1 {
			return nil
		}

		<-ctx.Done()
		return ctx.Err()
	}

	err := invoker.Race(context.Background(), f, f, f)
	require.NoError(err)
	require.Equal(uint64(3), atomic.LoadUint64(&count))
}

// Test adding tasks during execution.
func TestRaceAdd(t *testing.T) {
	require := require.New(t)

	tasks := invoker.New()
	count := uint64(0)

	f2 := func(ctx context.Context) (err error) {
		atomic.AddUint64(&count, 1)
		return nil
	}

	f := func(ctx context.Context) (err error) {
		atomic.AddUint64(&count, 1)

		// Each invocation adds another invocation that will return (and thus cancel).
		tasks.Add(f2)

		<-ctx.Done()
		return ctx.Err()
	}

	tasks.Add(f, f, f)

	err := tasks.Race(context.Background())
	require.NoError(err)
	require.Equal(uint64(6), atomic.LoadUint64(&count))
}

// Test queueing up tasks prior to running.
func TestRaceAddPending(t *testing.T) {
	require := require.New(t)

	count := uint64(0)
	f := func(ctx context.Context) (err error) {
		atomic.AddUint64(&count, 1)
		return nil
	}

	tasks := invoker.New()
	tasks.Add(f, f, f)
	tasks.Add(f, f)

	err := tasks.Race(context.Background())
	require.NoError(err)
	require.Equal(uint64(5), atomic.LoadUint64(&count))
}

// Test reusing the invoker object.
func TestRaceReuse(t *testing.T) {
	require := require.New(t)

	count := uint64(0)
	f := func(ctx context.Context) (err error) {
		atomic.AddUint64(&count, 1)
		return nil
	}

	tasks := invoker.New(f, f, f)

	err := tasks.Race(context.Background())
	require.NoError(err)
	require.Equal(uint64(3), atomic.LoadUint64(&count))

	err = tasks.Race(context.Background())
	require.Error(err)
}

// Test that two invoker objects can't be run at the same time.
func TestRaceRunning(t *testing.T) {
	require := require.New(t)

	f := func(ctx context.Context) (err error) {
		<-ctx.Done()
		return ctx.Err()
	}

	tasks := invoker.New(f, f, f)

	ctx, cancel := context.WithCancel(context.Background())
	errs := make(chan error, 2)

	go func() {
		errs <- tasks.Race(ctx)
		cancel()
	}()

	errs <- tasks.Race(ctx)
	cancel()

	require.Equal(invoker.ErrRunning, <-errs)
	require.Equal(context.Canceled, <-errs)
}
