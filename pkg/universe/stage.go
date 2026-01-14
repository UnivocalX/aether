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

type StageFn[T any] func(ctx context.Context, in T) Envelope[T]

type Stage[T any] struct {
	fn StageFn[T]
}

func NewStage[T any](fn StageFn[T]) *Stage[T] {
	return &Stage[T]{fn: fn}
}

func (s *Stage[T]) Run(
	ctx context.Context,
	stream <-chan Envelope[T],
) <-chan Envelope[T] {
	out := make(chan Envelope[T])

	go func() {
		defer close(out)

		for env := range stream {
			if env.Err != nil {
				select {
				case <-ctx.Done():
					return
				case out <- env:
				}
				continue
			}

			select {
			case <-ctx.Done():
				return
			case out <- s.fn(ctx, env.Value):
			}
		}
	}()

	return out
}
