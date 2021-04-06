package invoker

import (
	"context"
	"time"
)

// Sleep returns a Task that blocks until the given duration has passed.
func Sleep(d time.Duration) (t Task) {
	return func(ctx context.Context) (err error) {
		timer := time.NewTimer(d)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return nil
		}
	}
}
