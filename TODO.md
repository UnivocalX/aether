package universe

import "context"

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


package universe

import "context"

// Sink consumes a stream and performs an action.
func Sink[T any](
	ctx context.Context,
	in <-chan Envelope[T],
	fn func(Envelope[T]),
) error

// AbortOnError cancels the context on first error encountered.
func AbortOnError[T any](
	ctx context.Context,
	cancel context.CancelFunc,
	in <-chan Envelope[T],
) <-chan Envelope[T]

