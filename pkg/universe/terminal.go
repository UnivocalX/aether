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

type Consume[T any] func(Envelope[T]) error

func Sink[T any](
	ctx context.Context,
	fn Consume[T],
	stream <-chan Envelope[T],
) error {
	for env := range OrDone(ctx, stream) {
		if err := fn(env); err != nil {
			return err
		}
	}

	return ctx.Err()
}
