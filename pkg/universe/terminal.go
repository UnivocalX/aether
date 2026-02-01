package universe

import (
	"context"
)

type Consumer[T any] func(Envelope[T]) error
type Reducer[T, R any] func(R, Envelope[T]) R

// Consume reads envelopes from the stream and applies the consumer
// function to each one. This is intended for side‑effect processing.
// The function stops early if the consumer returns an error or if the
// context is
func Consume[T any](ctx context.Context, stream <-chan Envelope[T], fn Consumer[T]) error {
	for env := range OrDone(ctx, stream) {
		if err := fn(env); err != nil {
			return err
		}
	}
	return ctx.Err()
}

// Reduce folds all envelopes in the stream into a single accumulated result
// using the reducer function. The reduction stops immediately when an
// envelope carries an error. If the context is canceled, the partially
// reduced value and ctx.Err() are returned.
func Reduce[T, R any](ctx context.Context, stream <-chan Envelope[T], fn Reducer[T, R], init R) (R, error) {
	out := init

	for env := range OrDone(ctx, stream) {
		if env.Err != nil {
			return out, env.Err
		}
		out = fn(out, env)
	}

	return out, ctx.Err()
}

// Drain consumes and discards all envelopes from the stream.
// Useful when you need to exhaust a channel but don’t care about
// the values. Stops when the stream is closed or the context expires.
func Drain[T any](ctx context.Context, stream <-chan Envelope[T]) error {
	for range OrDone(ctx, stream) {
		// intentionally discard
	}
	return ctx.Err()
}

// Collect reads envelopes from the stream and appends their values to a slice.
// It stops on the first envelope error, when the slice reaches the given
// capacity, or when the context is canceled.
func Collect[T any](ctx context.Context, stream <-chan Envelope[T], cap int) ([]T, error) {
	out := make([]T, 0, cap)

	for env := range OrDone(ctx, stream) {
		if env.Err != nil {
			return out, env.Err
		}
		out = append(out, env.Value)

		if len(out) == cap {
			break
		}
	}

	return out, ctx.Err()
}

// Count returns the number of successfully received envelopes in the stream.
// It stops on the first envelope error or when context cancellation occurs.
func Count[T any](ctx context.Context, stream <-chan Envelope[T]) (int, error) {
	count := 0

	for env := range OrDone(ctx, stream) {
		if env.Err != nil {
			return count, env.Err
		}
		count++
	}

	return count, ctx.Err()
}

func Partition[T any](ctx context.Context, stream <-chan Envelope[T]) (success, failure []Envelope[T], err error) {
	for env := range OrDone(ctx, stream) {
		if env.Err != nil {
			failure = append(failure, env)
		} 	else {
			success = append(success, env)
		}
	}

	return success, failure, ctx.Err()
}
