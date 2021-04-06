package invoker

import (
	"context"
	"time"
)

// Return a Task that runs for the given amount of time before erroring.
func Timeout(duration time.Duration) Task {
	return func(ctx context.Context) (err error) {
		ctx, cancel := context.WithTimeout(ctx, duration)
		defer cancel()

		<-ctx.Done()
		return ctx.Err()
	}
}

// Return a Task that runs for the given amount of time before returning nil.
func Timer(duration time.Duration) Task {
	return func(ctx context.Context) (err error) {
		timer := time.NewTimer(duration)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return nil
		}
	}
}
