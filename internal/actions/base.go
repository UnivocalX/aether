package actions

import (
	"context"
	"runtime"
	"sync"
	"time"
)

const (
	DEFAULT_QUEUE_SIZE = 100
	DEFAULT_TIMEOUT    = 10 * time.Minute
)

type Runner interface {
	Run() error
}

type Action struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// NewAction creates a new Action with optional timeout
func NewAction(ctx context.Context, timeout ...time.Duration) *Action {
	if ctx == nil {
		ctx = context.Background()
	}

	t := DEFAULT_TIMEOUT
	if len(timeout) > 0 && timeout[0] > 0 {
		t = timeout[0]
	}

	c, cancel := context.WithTimeout(ctx, t)

	return &Action{
		ctx:    c,
		cancel: cancel,
	}
}

// Allow explicit cleanup if needed
func (act *Action) Cancel() {
	if act.cancel != nil {
		act.cancel()
	}
}

func (act *Action) CalculateNumOfRoutines(totalOperations int) int {
	if totalOperations <= 0 {
		return 1
	}

	numCPU := runtime.NumCPU()

	var totalRoutines int
	switch {
	case totalOperations <= 10:
		totalRoutines = 2
	case totalOperations <= 100:
		totalRoutines = numCPU
	case totalOperations <= 1000:
		totalRoutines = numCPU * 2
	default:
		totalRoutines = numCPU * 4
	}

	if totalRoutines > totalOperations {
		totalRoutines = totalOperations
	}

	return totalRoutines
}

func (act *Action) governedOPS(
	wg *sync.WaitGroup,
	fn func(item interface{}) error,
	queue chan interface{},
	errors chan error,
) {
	defer wg.Done()

	for {
		select {
		case <-act.ctx.Done():
			return
		case item, ok := <-queue:
			if !ok {
				return
			}
			if err := fn(item); err != nil {
				select {
				case errors <- err:
				case <-act.ctx.Done():
					return
				}
			}
		}
	}
}

// Parallel executes fn for each item in queue with timeout support.
func (act *Action) Parallel(fn func(item interface{}) error, queue chan interface{}) chan error {
	queueSize := cap(queue)
	if queueSize == 0 {
		queueSize = DEFAULT_QUEUE_SIZE
	}

	totalRoutines := act.CalculateNumOfRoutines(queueSize)
	errors := make(chan error, queueSize)

	var wg sync.WaitGroup
	wg.Add(totalRoutines)

	for i := 0; i < totalRoutines; i++ {
		go act.governedOPS(&wg, fn, queue, errors)
	}

	go func() {
		wg.Wait()
		close(errors)
	}()

	return errors
}