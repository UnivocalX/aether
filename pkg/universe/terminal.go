package universe

import "context"

func Drain[T any](
	ctx context.Context,
	stream <-chan T,
) error {
	for range OrDone(ctx, stream) {
		// intentionally discard
	}
	return ctx.Err()
}

func Collect[T any](
	ctx context.Context,
	stream <-chan T,
	cap int,
) ([]T, error) {
	out := make([]T, 0, cap)

	for value := range OrDone(ctx, stream) {
		out = append(out, value)

		if len(out) == cap {
			break
		}
	}

	return out, ctx.Err()
}

type Consumer[T any] func(Envelope[T]) error

func ForEach[T any](
	ctx context.Context,
	fn Consumer[T],
	stream <-chan Envelope[T],
) error {
	for env := range OrDone(ctx, stream) {
		if err := fn(env); err != nil {
			return err
		}
	}
	return ctx.Err()
}

type Reducer[T, R any] func(R, Envelope[T]) R

// Aggregate works with Envelope streams, handling errors appropriately
func Aggregate[T, R any](
	ctx context.Context,
	fn Reducer[T, R],
	stream <-chan Envelope[T],
	init R,
) (R, error) {
	out := init

	for env := range OrDone(ctx, stream) {
		if env.Err != nil {
			// Return the error immediately, along with current result
			return out, env.Err
		}
		out = fn(out, env)
	}

	return out, ctx.Err()
}

// Count returns the count of successful values in an Envelope stream
func Count[T any](
	ctx context.Context,
	stream <-chan Envelope[T],
) (int, error) {
	count := 0

	for env := range OrDone(ctx, stream) {
		if env.Err != nil {
			return count, env.Err
		}
		count++
	}

	return count, ctx.Err()
}
