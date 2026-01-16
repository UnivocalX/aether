package universe

import "context"

func Source[T any](ctx context.Context, values ...T) <-chan Envelope[T] {
	out := make(chan Envelope[T], len(values))

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

	return out
}

type Envelope[T any] struct {
	Value T
	Err   error
}

type Pipeline[T any] struct {
	operators []Operator[T]
}

func NewPipeline[T any](ops ...Operator[T]) *Pipeline[T] {
	return &Pipeline[T]{operators: ops}
}

func (p *Pipeline[T]) Run(
	ctx context.Context,
	source <-chan Envelope[T],
) <-chan Envelope[T] {

	stream := source
	for _, op := range p.operators {
		stream = op(ctx, stream)
	}
	return stream
}