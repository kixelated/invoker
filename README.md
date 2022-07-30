# ![](docs/invoker.png) invoker
Invoker provides a way to run goroutines in an easy and safe manner.

[documentation](https://pkg.go.dev/github.com/kixelated/invoker)

## Concepts

The `go` keyword is amazing because it makes it trivial to create and run new threads with minimal overhead. However, a lot of developers abuse the keyword, creating a mess of concurrency that needs to be manually managed. Fire-and-forget causes problems.

One of the concepts behind invoker is that functions should block on spawned goroutines. This is combined with contexts, which allow you to cancel spawned work to avoid blocking forever. When you write your code in this manner, concurrency becomes much easier.

## Core
The core unit of work is a `Task`, which is really just: `type Task = func(context.Context) (error)`

When you want to run multiple tasks, you provide them as arguments to either the `Run`, `Race`, or `Repeat` method. All of these methods will run tasks to completion, canceling the context and returning any errors depending on the desired behavior:

* `Run` will return the first error, or `nil` when all tasks have finished.
* `Race` will return the first result.
* `Repeat` will return the first error.

### Example
```go
// Make a habit of creating functions that take a context and return an error!
func Run(ctx context.Context) (err error) {
	// Returns a server with the method: Run(context.Context) (error)
	server := NewServer()
	
	// Returns a Task that runs for 5 minutes, then it errors
	waitTimeout := invoker.Timeout(5 * time.Minute)
	
	// Returns a Task that runs until an interrupt signal is received
	waitSignal := invoker.Interrupt
	
	// A Task is just any function that takes a context and returns an error
	printHello := func(ctx context.Context) (err error) {
		fmt.Println("hello world!")
		return nil
	}

	// invoker.Run will run all of the tasks in parallel and block until they all return.
	// On an error, the context is canceled and the error returned (once all functions have returned).
	return invoker.Run(ctx, server.Run, waitTimeout, waitSignal, printHello)
}
```

## Dynamic Tasks
Invoker supports dynamic `Tasks`, allowing you to `Add` a `Task` to an existing or future `Run`/`Race`/`Repeat` call.

```go
func runServer(ctx context.Context) (err error) {
	// returns a server with the methods:
	//   Run(context.Context) (error)
	//   Accept(context.Context) (Connection, erreror)
	server := NewServer()

	// Create the Tasks object that we'll use for all incoming connections
	var conns invoker.Tasks
  
	// Create a new task that will accept all incoming connections and make sure Run is called.
	accept := func(context.Context) (err error) {
		for {
			// Returns a connection object with the method:
			//   Run(context.Context) (error)
			conn, err := server.Accept(ctx)
			if err != nil {
				return err
			}

			// Immediately call the Run method on the connection to handle any per-connection state.
			// NOTE: in this example, any connection errors will terminate the server.
			// You probably want a wrapper to log errors and return `nil` instead.
			tasks.Add(conn.Run)
		}
	}

	// We run the server, our accept loop, and all accepted connections.
	// If any one of these functions returns an error, the others are cancelled.
	// NOTE: `Repeat` is used for `conns` such that it doesn't exit when there are no outstanding connections.
	return invoker.Run(ctx, server.Run, accept, conns.Repeat)
}
```

## Helpers
There are a few helper methods that create common `Task`s.

* `Signal(...os.Signal)` blocks until the provided signals are caught, and returns an `ErrSignal` error.
* `Interrupt` is short-hand for `Signal(syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)`.
* `Timeout(time.Duration)` blocks for the given duration and then returns `context.ErrTimeout`.
* `Timer(time.Duration)` blocks for the given duration and then returns `nil`.
* `Sleep(time.Duration)` is the same as `Timer`.
* `Context(context.Context)` blocks until an existing context is done.
* `Noop` does nothing!

## Panics
Invoker will run each `Task` in it's own Goroutine. Due to the lack of the `go` keyword in the invoker API, it's often difficult to realize that a panicing `Task` might crash the application.

To get the stack trace of the panic, use:
```golang
var errPanic invoker.ErrPanic
if errors.As(err, errPanic) {
	// output of debug.Stack()
	stack := errPanic.Stack()
}
```

By default, Invoker will catch any panics and wrap them in a `ErrPanic` object that includes the stack trace. You can disable this behavior with `invoker.Panic = true`. Note that this is a global setting should it should only be used for debugging.

## ErrGroup
Invoker is very similar to [errgroup](https://godoc.org/golang.org/x/sync/errgroup), but with an API designed for contexts. Invoker includes all of the extra functionality as mentioned above while using one fewer goroutine than errgroup. Here's the example code written with errgroup using the unwieldy API:

```go
// errgroup has GROSS context support
func Run(ctx context.Context) (err error) {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() (err error) {
		return server.Run(ctx)
	})

	g.Go(func() (err error) {
		// You have to write this using time.Timer or context.WithTimeout
		return waitTimeout(ctx)
	})

	g.Go(func() (err error) {
		// You have to write this using os.Signal
		return waitSignal(ctx)
	})
	
	g.Go(func() (err error) {
		return printHello(ctx)
	})

	return g.Wait()
}
```
