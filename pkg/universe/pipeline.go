package universe

import "context"

// Pipeline represents a composable stream of envelopes
type Pipeline[T any] struct {
	source func(context.Context) <-chan Envelope[T]
}

type Envelope[T any] struct {
	Value T
	Err   error
}

type Transformer[T, U any] func(Envelope[T]) Envelope[U]

type Predicate[T any] func(Envelope[T]) bool

type Observer[T any] func(Envelope[T])

// Of creates a Stream from values
func Of[T any](ctx context.Context, values ...T) *Pipeline[T] {
	outputStream := make(chan Envelope[T], len(values))

	go func() {
		defer close(outputStream)

		for _, v := range values {
			select {
			case <-ctx.Done():
				return
			case outputStream <- Envelope[T]{Value: v}:
			}
		}
	}()

	return &Pipeline[T]{
		source: func(context.Context) <-chan Envelope[T] {
			return outputStream
		},
	}
}

// Map applies a transformation that can change the type from T to U
// If workers > 1, the transformation runs concurrently
func Map[T, U any](s *Pipeline[T], fn Transformer[T, U], workers int) *Pipeline[U] {
	return &Pipeline[U]{
		source: func(ctx context.Context) <-chan Envelope[U] {
			worker := func(ctx context.Context, stream <-chan Envelope[T]) <-chan Envelope[U] {
				return transformWorker(ctx, stream, fn)
			}

			return FanIn(ctx, FanOut(ctx, worker, s.source(ctx), workers)...)
		},
	}
}

// Transform applies a transformation that keeps the same type T
// If workers > 1, the transformation runs concurrently
func (s *Pipeline[T]) Transform(fn Transformer[T, T], workers int) *Pipeline[T] {
	return &Pipeline[T]{
		source: func(ctx context.Context) <-chan Envelope[T] {
			worker := func(ctx context.Context, input <-chan Envelope[T]) <-chan Envelope[T] {
				return transformWorker(ctx, input, fn)
			}

			return FanIn(ctx, FanOut(ctx, worker, s.source(ctx), workers)...)
		},
	}
}

func transformWorker[T, U any](ctx context.Context, stream <-chan Envelope[T], fn Transformer[T, U]) <-chan Envelope[U] {
	outputStream := make(chan Envelope[U])

	go func() {
		defer close(outputStream)

		for env := range stream {
			select {
			case <-ctx.Done():
				return
			case outputStream <- fn(env):
			}
		}
	}()

	return outputStream
}

// Filter keeps only envelopes that match the predicate
// If workers > 1, filtering runs concurrently
func (s *Pipeline[T]) Filter(fn Predicate[T], workers int) *Pipeline[T] {
	return &Pipeline[T]{
		source: func(ctx context.Context) <-chan Envelope[T] {
			worker := func(ctx context.Context, stream <-chan Envelope[T]) <-chan Envelope[T] {
				return filterWorker(ctx, stream, fn)
			}

			return FanIn(ctx, FanOut(ctx, worker, s.source(ctx), workers)...)
		},
	}
}

func filterWorker[T any](ctx context.Context, stream <-chan Envelope[T], fn Predicate[T]) <-chan Envelope[T] {
	out := make(chan Envelope[T])

	go func() {
		defer close(out)

		for env := range stream {
			if fn(env) {
				select {
				case <-ctx.Done():
					return
				case out <- env:
				}
			}
		}
	}()

	return out
}

// Tap performs a side-effect on each envelope
// If workers > 1, side-effects run concurrently
func (s *Pipeline[T]) Tap(fn Observer[T], workers int) *Pipeline[T] {
	return &Pipeline[T]{
		source: func(ctx context.Context) <-chan Envelope[T] {
			worker := func(ctx context.Context, stream <-chan Envelope[T]) <-chan Envelope[T] {
				return tapWorker(ctx, stream, fn)
			}

			return FanIn(ctx, FanOut(ctx, worker, s.source(ctx), workers)...)
		},
	}
}

func tapWorker[T any](ctx context.Context, stream <-chan Envelope[T], fn Observer[T]) <-chan Envelope[T] {
	out := make(chan Envelope[T])

	go func() {
		defer close(out)
		for env := range stream {
			fn(env) // side-effect

			select {
			case <-ctx.Done():
				return
			case out <- env:
			}
		}
	}()

	return out
}

// UntilDone ensures the stream respects context cancellation
func (s *Pipeline[T]) UntilDone() *Pipeline[T] {
	return &Pipeline[T]{
		source: func(ctx context.Context) <-chan Envelope[T] {
			return OrDone(ctx, s.source(ctx))
		},
	}
}

// Run executes the stream and returns the output channel
func (s *Pipeline[T]) Run(ctx context.Context) <-chan Envelope[T] {
	return s.source(ctx)
}
