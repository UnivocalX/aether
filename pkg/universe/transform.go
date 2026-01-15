package universe

import (
	"context"

)

// Envelope represents a value flowing through the pipeline together
// with an optional error produced by a stage.
type Envelope[T any] struct {
	Value T
	Err   error
}

type Operator[T any] func(ctx context.Context, stream <-chan Envelope[T]) <-chan Envelope[T]

type Transformer[T any] func(input T) (T, error)

func Map[T any](
	fn Transformer[T],
) Operator[T] {
	return func(ctx context.Context, stream <-chan Envelope[T]) <-chan Envelope[T] {
		out := make(chan Envelope[T])

		go func() {
			defer close(out)
			
			for env := range OrDone(ctx, stream) {
				v, err := fn(env.Value)
				select {
				case <-ctx.Done():
					return
				case out <- Envelope[T]{Value: v, Err: err}:
				}
			}
		}()

		return out
	}
}

type Predicate[T any] func(T) bool

func Filter[T any](fn Predicate[T]) Operator[T] {
	return func(ctx context.Context, stream <-chan Envelope[T]) <-chan Envelope[T] {
		filtered := make(chan Envelope[T])

		go func ()  {
			defer close(filtered)

			for env := range OrDone(ctx, stream) {
				if fn(env.Value) {
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