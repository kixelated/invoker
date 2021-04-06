package invoker

import "fmt"

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
