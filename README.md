# ![](docs/invoker.png) invoker
Invoker provides a way to run goroutines in an easy and safe manner.

[documentation](https://pkg.go.dev/github.com/kixelated/invoker)

## Concepts

The `go` keyword is amazing because it makes it trivial to create and run new threads with minimal overhead. However, a lot of developers abuse the keyword, creating a mess of concurrency that needs to be manually managed. Fire-and-forget causes problems.

One of the concepts behind invoker is that functions should block on spawned goroutines. This is combined with contexts, which allow you to cancel spawned work to avoid blocking forever. When you write your code in this manner, concurrency becomes much easier.

## Example
```go
func Run(ctx context.Context) (err error) {
	// invoker.Run will run all of the tasks in parallel and block until they all return.
	// On an error, the context is canceled and the error returned (once all functions have returned).
	return invoker.Run(ctx, runServer, foldLaundry, waitSignal)
}

func runServer(ctx context.Context) (err error) {
	// ...you could do something long-lived like run a HTTP server.
}

func foldLaundry(ctx context.Context) (err error) {
	// ...or you could do something short-lived and return nil when finished.
}

func waitSignal(ctx context.Context) (err error) {
	// ...or you could just wait for an os.Signal, which will gracefully cancel and wait for goroutines to return.
}
```

## ErrGroup
Invoker is very similar to [errgroup](https://godoc.org/golang.org/x/sync/errgroup), but with an API designed for contexts. Here's the same `Run` function above but written with errgroup:

```go
// errgroup has GROSS context support
func Run(ctx context.Context) (err error) {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() (err error) {
		return runServer(ctx)
	})

	g.Go(func() (err error) {
		return foldLaundry(ctx)
	})

	g.Go(func() (err error) {
		return waitSignal(ctx)
	})

	return g.Wait()
}
```

Invoker has some additional functionality:
* `Run` will return the first non-nil result, or nil when all tasks have finished.
* `Race` will return the first result.
* `Repeat` will result the first non-nil result.
* `New` creates a `invoker.Tasks` object, which supports the above and can `Add` more tasks.
* Uses one fewer goroutine than errgroup.
