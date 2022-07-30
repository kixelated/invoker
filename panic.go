package invoker

import "fmt"

// Set to true to not catch panics, primarily for debugging to easily print the stack trace.
// Your production application should use `errors.As(err, ErrPanic)` to print panics as errors.
var Panic bool = false

type ErrPanic struct {
	p     interface{}
	stack []byte
}

func (ep ErrPanic) Error() string {
	return fmt.Sprintf("caught panic: %v", ep.p)
}

func (ep ErrPanic) Stack() []byte {
	return ep.stack
}
