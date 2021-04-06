package invoker

import (
	"context"
)

// Task is a function that will execute until finished or the context is done.
type Task func(ctx context.Context) (err error)

// Run will execute the given tasks, returning the first error result and canceling any remaining tasks.
func Run(ctx context.Context, tasks ...Task) (err error) {
	return New(tasks...).Run(ctx)
}

// Race will execute the given tasks, returning the first result and canceling any remaining tasks.
func Race(ctx context.Context, tasks ...Task) (err error) {
	return New(tasks...).Race(ctx)
}

// Wait blocks until the context is canceled
func Wait(ctx context.Context) (err error) {
	<-ctx.Done()
	return ctx.Err()
}

// Noop returns immediately
func Noop(ctx context.Context) (err error) {
	return nil
}
