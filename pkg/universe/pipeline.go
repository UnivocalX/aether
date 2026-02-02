package universe

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
)

// Pipeline represents a composable, lazily-evaluated stream of envelopes.
type Pipeline[T any] struct {
	generator func(ctx context.Context) Stream[T]
}

// Stream wraps a channel of Envelope[T] and associated metadata.
type Stream[T any] struct {
	Meta *Meta
	Data <-chan Envelope[T]
}

// Envelope wraps a value and an optional error.
type Envelope[T any] struct {
	Value T
	Err   error
}

// Meta holds attributes and tracks errors during pipeline execution.
type Meta struct {
	OriginTotalItems int
	Attributes       sync.Map
	errorOccurred    atomic.Bool
}

func (m *Meta) MarkErrorOccurred(statement bool) {
	if statement {
		m.errorOccurred.Store(true)
	}
}

func (m *Meta) ErrorOccurred() bool {
	return m.errorOccurred.Load()
}

// Transformer defines a function that transforms an Envelope[T] into an Envelope[U].
type Transformer[T, U any] func(*Meta, Envelope[T]) Envelope[U]

// Predicate defines a function to filter envelopes.
type Predicate[T any] func(*Meta, Envelope[T]) bool

// Observer defines a function to perform side-effects on envelopes.
type Observer[T any] func(*Meta, Envelope[T])

// From creates a Pipeline from a list of values.
func From[T any](ctx context.Context, values ...T) *Pipeline[T] {
	slog.Debug("creating pipeline from list of values", "totalItems", len(values))

	return &Pipeline[T]{
		generator: func(ctx context.Context) Stream[T] {
			out := make(chan Envelope[T])
			meta := &Meta{OriginTotalItems: len(values)}

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

			return Stream[T]{Data: out, Meta: meta}
		},
	}
}

// Map applies a transformation to each envelope in the pipeline with worker parallelism.
func Map[T, U any](p *Pipeline[T], fn Transformer[T, U], workers int) *Pipeline[U] {
	return &Pipeline[U]{
		generator: func(ctx context.Context) Stream[U] {
			inStream := p.generator(ctx)
			meta := inStream.Meta

			worker := func(ctx context.Context, in <-chan Envelope[T]) <-chan Envelope[U] {
				out := make(chan Envelope[U])
				go func() {
					defer close(out)
					for env := range in {
						oe := fn(meta, env)
						select {
						case <-ctx.Done():
							return
						case out <- oe:
							meta.MarkErrorOccurred(oe.Err != nil)
						}
					}
				}()
				return out
			}

			fanned := FanOut(ctx, worker, inStream.Data, workers)
			return Stream[U]{Data: FanIn(ctx, fanned...), Meta: meta}
		},
	}
}

// Transform applies a transformation that keeps the same type.
func (p *Pipeline[T]) Transform(fn Transformer[T, T], workers int) *Pipeline[T] {
	return &Pipeline[T]{
		generator: func(ctx context.Context) Stream[T] {
			inStream := p.generator(ctx)
			meta := inStream.Meta

			worker := func(ctx context.Context, in <-chan Envelope[T]) <-chan Envelope[T] {
				out := make(chan Envelope[T])
				go func() {
					defer close(out)
					for env := range in {
						oe := fn(meta, env)
						select {
						case <-ctx.Done():
							return
						case out <- oe:
							meta.MarkErrorOccurred(oe.Err != nil)
						}
					}
				}()
				return out
			}

			fanned := FanOut(ctx, worker, inStream.Data, workers)
			return Stream[T]{Data: FanIn(ctx, fanned...), Meta: meta}
		},
	}
}

// Filter passes through only envelopes that satisfy the predicate.
func (p *Pipeline[T]) Filter(fn Predicate[T], workers int) *Pipeline[T] {
	return &Pipeline[T]{
		generator: func(ctx context.Context) Stream[T] {
			inStream := p.generator(ctx)
			meta := inStream.Meta

			worker := func(ctx context.Context, in <-chan Envelope[T]) <-chan Envelope[T] {
				out := make(chan Envelope[T])
				go func() {
					defer close(out)
					for env := range in {
						if !fn(meta, env) {
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

			fanned := FanOut(ctx, worker, inStream.Data, workers)
			return Stream[T]{Data: FanIn(ctx, fanned...), Meta: meta}
		},
	}
}

// Tap executes a side-effect function for each envelope.
func (p *Pipeline[T]) Tap(fn Observer[T], workers int) *Pipeline[T] {
	return &Pipeline[T]{
		generator: func(ctx context.Context) Stream[T] {
			inStream := p.generator(ctx)
			meta := inStream.Meta

			worker := func(ctx context.Context, in <-chan Envelope[T]) <-chan Envelope[T] {
				out := make(chan Envelope[T])
				go func() {
					defer close(out)
					for env := range in {
						fn(meta, env)
						select {
						case <-ctx.Done():
							return
						case out <- env:
						}
					}
				}()
				return out
			}

			fanned := FanOut(ctx, worker, inStream.Data, workers)
			return Stream[T]{Data: FanIn(ctx, fanned...), Meta: meta}
		},
	}
}

// UntilDone ensures the pipeline respects context cancellation.
func (p *Pipeline[T]) UntilDone() *Pipeline[T] {
	return &Pipeline[T]{
		generator: func(ctx context.Context) Stream[T] {
			inStream := p.generator(ctx)
			return Stream[T]{Data: OrDone(ctx, inStream.Data), Meta: inStream.Meta}
		},
	}
}

// Run starts the pipeline and returns the final Stream.
func (p *Pipeline[T]) Run(ctx context.Context) Stream[T] {
	return p.generator(ctx)
}
