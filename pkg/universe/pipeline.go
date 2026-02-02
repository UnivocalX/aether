package universe

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
)

// Pipeline represents a composable, lazily‑evaluated stream of envelopes.
// Each Pipeline stage defines how envelopes are produced or transformed.
type Pipeline[T any] struct {
	generator func(context.Context) <-chan Envelope[T]
	Meta      *Meta
}

// Envelope wraps a value of type T and an optional error.
// Errors propagate through the pipeline but do not crash goroutines
type Stream[T any] struct {
	*Meta
	Data <-chan Envelope[T]
}

type Envelope[T any] struct {
	Value T
	Err   error
}

type Meta struct {
	Attributes       sync.Map
	OriginTotalItems int
	hasError         atomic.Bool
}

func (m *Meta) MarkHasErrors(statement bool) {
	if statement {
		m.hasError.Store(true)
	}
}

func (m *Meta) HasErrors() bool {
	return m.hasError.Load()
}

type Transformer[T, U any] func(*Meta, Envelope[T]) Envelope[U]
type Predicate[T any] func(*Meta, Envelope[T]) bool
type Observer[T any] func(*Meta, Envelope[T])

// From creates a Pipeline from a fixed list of values.
// Each value is emitted as an Envelope with no error.
// Emission stops early if the context is canceled.
func From[T any](ctx context.Context, values ...T) *Pipeline[T] {
	slog.Debug("creating pipeline from list of values", "totalItems", len(values))
	out := make(chan Envelope[T])

	go func() {
		defer close(out)

		for _, v := range values {
			select {
			case <-ctx.Done():
				return
			case out <- Envelope[T]{Value: v}:
			}
		}
	}()

	return &Pipeline[T]{
		Meta: &Meta{OriginTotalItems: len(values)},
		generator: func(context.Context) <-chan Envelope[T] {
			return out
		},
	}
}

// Map applies a transformation to each envelope.
func Map[T, U any](s *Pipeline[T], fn Transformer[T, U], workers int) *Pipeline[U] {
	return &Pipeline[U]{
		Meta: s.Meta,
		generator: func(ctx context.Context) <-chan Envelope[U] {
			worker := func(ctx context.Context, in <-chan Envelope[T]) <-chan Envelope[U] {
				out := make(chan Envelope[U])

				go func() {
					defer close(out)

					for env := range in {
						oe := fn(s.Meta, env)

						select {
						case <-ctx.Done():
							return
						case out <- oe:
							s.Meta.MarkHasErrors(oe.Err != nil)
						}
					}
				}()

				return out
			}

			return FanIn(ctx, FanOut(ctx, worker, s.generator(ctx), workers)...)
		},
	}
}

// Transform applies a transformation that keeps the same type.
func (p *Pipeline[T]) Transform(fn Transformer[T, T], workers int) *Pipeline[T] {
	return &Pipeline[T]{
		Meta: p.Meta,
		generator: func(ctx context.Context) <-chan Envelope[T] {
			worker := func(ctx context.Context, in <-chan Envelope[T]) <-chan Envelope[T] {
				out := make(chan Envelope[T])

				go func() {
					defer close(out)
					for env := range in {
						oe := fn(p.Meta, env)

						select {
						case <-ctx.Done():
							return
						case out <- oe:
							p.Meta.MarkHasErrors(oe.Err != nil)
						}
					}
				}()

				return out
			}

			return FanIn(ctx, FanOut(ctx, worker, p.generator(ctx), workers)...)
		},
	}
}

// Filter passes through only envelopes that satisfy the predicate.
func (p *Pipeline[T]) Filter(fn Predicate[T], workers int) *Pipeline[T] {
	return &Pipeline[T]{
		generator: func(ctx context.Context) <-chan Envelope[T] {
			worker := func(ctx context.Context, in <-chan Envelope[T]) <-chan Envelope[T] {
				out := make(chan Envelope[T])

				go func() {
					defer close(out)
					for env := range in {
						if !fn(p.Meta, env) {
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

			return FanIn(ctx, FanOut(ctx, worker, p.generator(ctx), workers)...)
		},
	}
}

// Tap executes a side-effect for each envelope.
func (p *Pipeline[T]) Tap(fn Observer[T], workers int) *Pipeline[T] {
	return &Pipeline[T]{
		generator: func(ctx context.Context) <-chan Envelope[T] {
			worker := func(ctx context.Context, in <-chan Envelope[T]) <-chan Envelope[T] {
				out := make(chan Envelope[T])

				go func() {
					defer close(out)
					for env := range in {
						fn(p.Meta, env)

						select {
						case <-ctx.Done():
							return
						case out <- env:
						}
					}
				}()

				return out
			}

			return FanIn(ctx, FanOut(ctx, worker, p.generator(ctx), workers)...)
		},
	}
}

// UntilDone wraps the pipeline so that it stops producing values as soon
// as the context is canceled. This helps ensure all stages respect cancellation.
func (p *Pipeline[T]) UntilDone() *Pipeline[T] {
	return &Pipeline[T]{
		generator: func(ctx context.Context) <-chan Envelope[T] {
			return OrDone(ctx, p.generator(ctx))
		},
	}
}

// Run starts the pipeline and returns the final output channel.
// It is the terminal operation that consumes all upstream definitions.
func (p *Pipeline[T]) Run(ctx context.Context) Stream[T] {
	return Stream[T]{
		Data: p.generator(ctx),
		Meta: p.Meta,
	}
}
