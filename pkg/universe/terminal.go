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

func Consume[T any](
	ctx context.Context,
	stream <-chan Envelope[T],
	fn Consumer[T],
) error {
	for env := range OrDone(ctx, stream) {
		if err := fn(env); err != nil {
			return err
		}
	}
	return ctx.Err()
}

// Reduce works with Envelope streams, handling errors appropriately
func Reduce[T, R any](
	ctx context.Context,
	stream <-chan Envelope[T],
	fn Reducer[T, R],
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
