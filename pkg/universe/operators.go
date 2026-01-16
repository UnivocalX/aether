package universe

import (
	"context"
)

type Operator[T any] func(ctx context.Context, stream <-chan Envelope[T]) <-chan Envelope[T]

type Transformer[T any] func(Envelope[T]) Envelope[T]

func Map[T any](
	fn Transformer[T],
) Operator[T] {
	return func(ctx context.Context, stream <-chan Envelope[T]) <-chan Envelope[T] {
		out := make(chan Envelope[T])

		go func() {
			defer close(out)

			for env := range OrDone(ctx, stream) {
				select {
				case <-ctx.Done():
					return
				case out <- fn(env):
				}
			}
		}()

		return out
	}
}

type Predicator[T any] func(Envelope[T]) bool

func Filter[T any](fn Predicator[T]) Operator[T] {
	return func(ctx context.Context, stream <-chan Envelope[T]) <-chan Envelope[T] {
		filtered := make(chan Envelope[T])

		go func() {
			defer close(filtered)

			for env := range OrDone(ctx, stream) {
				if fn(env) {
					select {
					case <-ctx.Done():
						return
					case filtered <- env:
					}
				}
			}
		}()

		return filtered
	}
}

type Observer[T any] func(Envelope[T])

func Tap[T any](fn Observer[T]) Operator[T] {
	return func(ctx context.Context, stream <-chan Envelope[T]) <-chan Envelope[T] {
		out := make(chan Envelope[T])

		go func() {
			defer close(out)
			for env := range OrDone(ctx, stream) {
				fn(env) // side-effect

				select {
				case <-ctx.Done():
					return
				case out <- env: // forward unchanged
				}
			}
		}()

		return out
	}
}

func Concurrency[T any](op Operator[T], workers int) Operator[T] {
	return func(ctx context.Context, stream <-chan Envelope[T]) <-chan Envelope[T] {
		return FanIn(ctx, FanOut(ctx, op, stream, workers)...)
	}
}

func UntilDone[T any]() Operator[T] {
	return func(ctx context.Context, stream <-chan Envelope[T]) <-chan Envelope[T] {
		return OrDone(ctx, stream)
	}
}
