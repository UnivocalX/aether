package universe

import "context"

type Consumer[T any] func(Envelope[T]) error

type Reducer[T, R any] func(R, Envelope[T]) R

// Consume processes each envelope with a consumer function
func Consume[T any](ctx context.Context, stream <-chan Envelope[T], fn Consumer[T]) error {
	for env := range OrDone(ctx, stream) {
		if err := fn(env); err != nil {
			return err
		}
	}
	return ctx.Err()
}

// Reduce folds the stream into a single value
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

// Drain discards all values from the stream
func Drain[T any](ctx context.Context, stream <-chan Envelope[T]) error {
	for range OrDone(ctx, stream) {
		// intentionally discard
	}
	return ctx.Err()
}

// Collect gathers all results from the stream
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

// Count returns the count of successful values in the stream
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
