package universe

import "context"

// Core abstraction (assumed to already exist)
type Envelope[T any] struct {
	Value T
	Err   error
	// optional: Meta, RetryCount, Timestamp, etc.
}

type Stage[T any] func(ctx context.Context, in <-chan Envelope[T]) <-chan Envelope[T]


// Map transforms each value in the stream (pure function).
func Map[T any](
	fn func(T) (T, error),
) Stage[T]


// Filter passes through only values matching the predicate.
func Filter[T any](
	predicate func(T) bool,
) Stage[T]


// Tap performs side-effects without modifying the stream.
func Tap[T any](
	fn func(Envelope[T]),
) Stage[T]


// Partition splits a stream into two based on a predicate.
// This is NOT a Stage because it changes topology.
func Partition[T any](
	ctx context.Context,
	in <-chan Envelope[T],
	predicate func(T) bool,
) (
	matched <-chan Envelope[T],
	rest <-chan Envelope[T],
)


// Retry retries failed envelopes up to maxRetries.
func Retry[T any](
	maxRetries int,
	shouldRetry func(err error) bool,
) Stage[T]


// Batch groups values into fixed-size slices.
func Batch[T any](
	size int,
) Stage[[]T]


// Window groups values into time-based windows.
func Window[T any](
	windowSize time.Duration,
) Stage[[]T]


// MergeSorted merges multiple ordered streams into one ordered stream.
func MergeSorted[T any](
	less func(a, b T) bool,
	streams ...<-chan Envelope[T],
) <-chan Envelope[T]


package universe

import "context"

// Sink consumes a stream and performs an action.
func Sink[T any](
	ctx context.Context,
	in <-chan Envelope[T],
	fn func(Envelope[T]),
) error


// Drain consumes and discards all values.
func Drain[T any](
	ctx context.Context,
	in <-chan Envelope[T],
)


// Collect gathers all envelopes into memory.
func Collect[T any](
	ctx context.Context,
	in <-chan Envelope[T],
) ([]Envelope[T], error)


// AbortOnError cancels the context on first error encountered.
func AbortOnError[T any](
	ctx context.Context,
	cancel context.CancelFunc,
	in <-chan Envelope[T],
) <-chan Envelope[T]


package universe

import "context"

type Pipeline[T any] struct {
	ctx    context.Context
	stages []Stage[T]
}


// NewPipeline creates a pipeline with a base context.
func NewPipeline[T any](
	ctx context.Context,
) *Pipeline[T]


// Then appends a stage sequentially.
func (p *Pipeline[T]) Then(
	stage Stage[T],
) *Pipeline[T]


// ThenConcurrent appends a stage with fan-out concurrency.
func (p *Pipeline[T]) ThenConcurrent(
	workers int,
	stage Stage[T],
) *Pipeline[T]


// Run executes the pipeline starting from a source stream.
func (p *Pipeline[T]) Run(
	source <-chan Envelope[T],
) <-chan Envelope[T]


// Collect executes the pipeline and gathers results.
func (p *Pipeline[T]) Collect(
	source <-chan Envelope[T],
) ([]Envelope[T], error)
