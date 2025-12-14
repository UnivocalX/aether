package actions

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

const (
	DEFAULT_TIMEOUT         = 10 * time.Minute
	DEFAULT_MIN_PARALLELISM = 2
	DEFAULT_MAX_PARALLELISM = -1
)

// Artifact represents the outcome of a single operation
type Artifact[In any, Out any] struct {
	Input In
	Value Out
	Err   error
}

func (r Artifact[In, Out]) Unwrap() (Out, In, error) {
	return r.Value, r.Input, r.Err
}

// Action provides concurrent execution with context support
// In = input type, Out = output type
type Action[In any, Out any] struct {
	ctx            context.Context
	cancel         context.CancelFunc
	maxParallelism int
	minParallelism int
}

// NewAction creates a new Action with default settings
func NewAction[In any, Out any]() *Action[In, Out] {
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_TIMEOUT)
	return &Action[In, Out]{
		ctx:            ctx,
		cancel:         cancel,
		minParallelism: DEFAULT_MIN_PARALLELISM,
		maxParallelism: DEFAULT_MAX_PARALLELISM,
	}
}

// SetContext sets a custom context
func (a *Action[In, Out]) SetContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	if a.cancel != nil {
		a.cancel()
	}
	a.ctx, a.cancel = context.WithCancel(ctx)
	return nil
}

// SetTimeout sets a timeout
func (a *Action[In, Out]) SetTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got: %v", timeout)
	}
	if a.cancel != nil {
		a.cancel()
	}
	a.ctx, a.cancel = context.WithTimeout(context.Background(), timeout)
	return nil
}

// SetMaxParallelism sets max concurrent workers (-1 for unlimited)
func (a *Action[In, Out]) SetMaxParallelism(max int) error {
	if max < -1 {
		return fmt.Errorf("maxParallelism must be -1 or greater, got: %d", max)
	}
	a.maxParallelism = max
	if a.maxParallelism > 0 && a.minParallelism > a.maxParallelism {
		a.minParallelism = a.maxParallelism
	}
	return nil
}

// SetMinParallelism sets min concurrent workers
func (a *Action[In, Out]) SetMinParallelism(min int) error {
	if min <= 0 {
		return fmt.Errorf("minParallelism must be positive, got: %d", min)
	}
	a.minParallelism = min
	if a.maxParallelism > 0 && a.minParallelism > a.maxParallelism {
		a.minParallelism = a.maxParallelism
	}
	return nil
}

// workers calculates optimal number of workers
func (a *Action[In, Out]) workers(total int) int {
	if total <= 0 {
		return 1
	}

	numCPU := runtime.NumCPU()
	var w int

	switch {
	case total <= 10:
		w = 2
	case total <= 100:
		w = numCPU
	case total <= 1000:
		w = numCPU * 2
	default:
		w = numCPU * 4
	}

	if w > total {
		w = total
	}
	if a.maxParallelism > 0 && w > a.maxParallelism {
		w = a.maxParallelism
	}
	if w < a.minParallelism {
		w = a.minParallelism
	}

	return w
}

// worker processes items from the queue and sends results to output
func worker[In any, Out any](
	ctx context.Context,
	wg *sync.WaitGroup,
	queue <-chan In,
	out chan<- Artifact[In, Out],
	fn func(In) (Out, error),
) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case item, ok := <-queue:
			if !ok {
				return
			}
			value, err := fn(item)
			result := Artifact[In, Out]{Input: item, Value: value, Err: err}
			select {
			case out <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}

// Run executes fn on each input in parallel, returns results channel
// fn signature: func(input In) (output Out, error)
func (a *Action[In, Out]) Run(fn func(In) (Out, error), inputs ...In) <-chan Artifact[In, Out] {
	out := make(chan Artifact[In, Out], len(inputs))

	if len(inputs) == 0 {
		close(out)
		return out
	}

	queue := make(chan In, len(inputs))
	for _, item := range inputs {
		queue <- item
	}
	close(queue)

	var wg sync.WaitGroup
	numWorkers := a.workers(len(inputs))
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go worker(a.ctx, &wg, queue, out, fn)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// Collect gathers all results from the channel into a slice
func (a *Action[In, Out]) Collect(ch <-chan Artifact[In, Out]) []Artifact[In, Out] {
	var results []Artifact[In, Out]
	for result := range ch {
		results = append(results, result)
	}
	return results
}

// Cancel cancels the action
func (a *Action[In, Out]) Cancel() {
	if a.cancel != nil {
		a.cancel()
	}
}

// Context returns the action's context
func (a *Action[In, Out]) Context() context.Context {
	return a.ctx
}
