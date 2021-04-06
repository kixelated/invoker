package invoker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Signal returns a Task that blocks until one of the given signals is triggered.
func Signal(signals ...os.Signal) (t Task) {
	return func(ctx context.Context) (err error) {
		c := make(chan os.Signal, 1)

		signal.Notify(c, signals...)
		defer signal.Stop(c)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case sig := <-c:
			return ErrSignal{sig: sig}
		}
	}
}

// Interrupt is a Task that blocks until a terminate signal.
// Specifically: SIGTERM (kill default), SIGINT (ctrl+c), and SIGHUP (common kill signal)
func Interrupt(ctx context.Context) (err error) {
	return Signal(syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)(ctx)
}

// ErrSignal is returned with the signal recieved.
type ErrSignal struct {
	sig os.Signal
}

func (es ErrSignal) Error() string {
	return fmt.Sprintf("recieved signal: %s", es.sig)
}

func (es ErrSignal) Signal() os.Signal {
	return es.sig
}
