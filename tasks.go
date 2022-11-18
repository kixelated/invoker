package invoker

import (
	"context"
	"fmt"
	"sync"
)

// ErrRunning is returned if two goroutine try to similtaniously call Run/Race.
var ErrRunning = fmt.Errorf("already running")
var ErrFinished = fmt.Errorf("finished execution")

type mode int

const (
	modeInit mode = iota
	modeRun
	modeRace
	modeRepeat
	modeDone
)

// Tasks allows you to add additional tasks during a Run/Race.
type Tasks struct {
	mutex sync.Mutex

	mode    mode
	pending []Task

	running int
	first   bool
	err     error

	ctx    context.Context
	cancel context.CancelFunc
	done   chan error
}

// New constructs an Tasks instance allowing you to run additional tasks.
func New(tasks ...Task) (ts *Tasks) {
	ts = new(Tasks)
	ts.pending = tasks
	return ts
}

// Adds tasks to be executed.
// If Run has already completed, the tasks are executed but immediately cancelled.
func (ts *Tasks) Add(tasks ...Task) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if ts.mode == modeInit {
		ts.pending = append(ts.pending, tasks...)
		return
	}

	ts.running += len(tasks)

	for _, t := range tasks {
		go ts.run(ts.ctx, t)
	}
}

// Run returns the first error result (if any) and cancels any remaining tasks.
func (ts *Tasks) Run(ctx context.Context) (err error) {
	return ts.do(ctx, modeRun)
}

// Race returns the first result and cancels any remaining tasks.
func (ts *Tasks) Race(ctx context.Context) (err error) {
	return ts.do(ctx, modeRace)
}

// Repeat returns the first error result and cancels any remaining tasks.
func (ts *Tasks) Repeat(ctx context.Context) (err error) {
	return ts.do(ctx, modeRepeat)
}

func (ts *Tasks) do(ctx context.Context, m mode) (err error) {
	ts.mutex.Lock()

	switch ts.mode {
	case modeInit:
		// expected
	case modeDone:
		ts.mutex.Unlock()
		return ErrFinished
	default:
		ts.mutex.Unlock()
		return ErrRunning
	}

	tasks := ts.pending
	ts.pending = nil

	// If there are no tasks, advance to done directly.
	if len(tasks) == 0 && m != modeRepeat {
		ts.mode = modeDone
		ts.mutex.Unlock()
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ts.mode = m
	ts.ctx = ctx
	ts.cancel = cancel
	ts.first = true
	ts.running = len(tasks)

	ts.done = make(chan error, 1)
	ts.mutex.Unlock()

	for _, f := range tasks {
		go ts.run(ctx, f)
	}

	// Wait until all goroutines have exited
	return <-ts.done
}

func (ts *Tasks) run(ctx context.Context, t Task) {
	err := t(ctx)
	ts.report(err)
}

func (ts *Tasks) report(err error) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	ts.running -= 1

	switch ts.mode {
	case modeRun, modeRepeat:
		if ts.err == nil {
			ts.err = err
		}

		if err != nil {
			ts.cancel()
		}
	case modeRace:
		if ts.first {
			ts.err = err
			ts.first = false
		}

		ts.cancel()
	case modeDone:
		// already done
		return
	}

	if ts.running > 0 {
		return
	}

	// If it's the repeat mode, make sure there's an error before we stop
	if ts.mode != modeRepeat || ts.err != nil {
		// NOTE: This will be written to exactly once.
		ts.done <- ts.err
		ts.mode = modeDone
	}
}
