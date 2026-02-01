package universe

import (
	"context"
	"log/slog"
)

// Pipeline represents a composable, lazily‑evaluated stream of envelopes.
// Each Pipeline stage defines how envelopes are produced or transformed.
type Pipeline[T any] struct {
	source func(context.Context) <-chan Envelope[T]
}

// Envelope wraps a value of type T and an optional error.
// Errors propagate through the pipeline but do not crash goroutines
type Envelope[T any] struct {
	Value T
	Err   error
}

type Transformer[T, U any] func(Envelope[T]) Envelope[U]
type Predicate[T any] func(Envelope[T]) bool
type Observer[T any] func(Envelope[T])

// From creates a Pipeline from a fixed list of values.
// Each value is emitted as an Envelope with no error.
// Emission stops early if the context is canceled.
func From[T any](ctx context.Context, values ...T) *Pipeline[T] {
	slog.Debug("creating pipeline from list of values", "totalItems", len(values))
	outputStream := make(chan Envelope[T])

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

// Map applies a transformation to each envelope.
func Map[T, U any](s *Pipeline[T], fn Transformer[T, U], workers int) *Pipeline[U] {
	return &Pipeline[U]{
		source: func(ctx context.Context) <-chan Envelope[U] {
			worker := func(ctx context.Context, in <-chan Envelope[T]) <-chan Envelope[U] {
				out := make(chan Envelope[U])

				go func() {
					defer close(out)
					for env := range in {
						select {
						case <-ctx.Done():
							return
						case out <- fn(env):
						}
					}
				}()

				return out
			}

			return FanIn(ctx, FanOut(ctx, worker, s.source(ctx), workers)...)
		},
	}
}

// Transform applies a transformation that keeps the same type.
func (s *Pipeline[T]) Transform(fn Transformer[T, T], workers int) *Pipeline[T] {
	return &Pipeline[T]{
		source: func(ctx context.Context) <-chan Envelope[T] {
			worker := func(ctx context.Context, in <-chan Envelope[T]) <-chan Envelope[T] {
				out := make(chan Envelope[T])

				go func() {
					defer close(out)
					for env := range in {
						select {
						case <-ctx.Done():
							return
						case out <- fn(env):
						}
					}
				}()

				return out
			}

			return FanIn(ctx, FanOut(ctx, worker, s.source(ctx), workers)...)
		},
	}
}

// Filter passes through only envelopes that satisfy the predicate.
func (s *Pipeline[T]) Filter(fn Predicate[T], workers int) *Pipeline[T] {
	return &Pipeline[T]{
		source: func(ctx context.Context) <-chan Envelope[T] {
			worker := func(ctx context.Context, in <-chan Envelope[T]) <-chan Envelope[T] {
				out := make(chan Envelope[T])

				go func() {
					defer close(out)
					for env := range in {
						if !fn(env) {
							continue
						}
						select {
						case <-ctx.Done():
							return
						case out <- env:
						}
					}
				}()

				return out
			}

			return FanIn(ctx, FanOut(ctx, worker, s.source(ctx), workers)...)
		},
	}
}

// Tap executes a side-effect for each envelope.
func (s *Pipeline[T]) Tap(fn Observer[T], workers int) *Pipeline[T] {
	return &Pipeline[T]{
		source: func(ctx context.Context) <-chan Envelope[T] {
			worker := func(ctx context.Context, in <-chan Envelope[T]) <-chan Envelope[T] {
				out := make(chan Envelope[T])

				go func() {
					defer close(out)
					for env := range in {
						fn(env)

						select {
						case <-ctx.Done():
							return
						case out <- env:
						}
					}
				}()

				return out
			}

			return FanIn(ctx, FanOut(ctx, worker, s.source(ctx), workers)...)
		},
	}
}

// UntilDone wraps the pipeline so that it stops producing values as soon
// as the context is canceled. This helps ensure all stages respect cancellation.
func (s *Pipeline[T]) UntilDone() *Pipeline[T] {
	return &Pipeline[T]{
		source: func(ctx context.Context) <-chan Envelope[T] {
			return OrDone(ctx, s.source(ctx))
		},
	}
}

// Run starts the pipeline and returns the final output channel.
// It is the terminal operation that consumes all upstream definitions.
func (s *Pipeline[T]) Run(ctx context.Context) <-chan Envelope[T] {
	return s.source(ctx)
}
