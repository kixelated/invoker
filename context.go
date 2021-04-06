package invoker

import "context"

// Returns a Task that waits on the given context.
// Thus, this can used to wait on two contexts.
func Context(ctx context.Context) Task {
	return func(ctx2 context.Context) (err error) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ctx2.Done():
			return ctx2.Err()
		}
	}
}
