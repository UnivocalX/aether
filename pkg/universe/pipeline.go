package universe

import "context"

type Pipeline[T any] struct {
	ctx    context.Context
	stream <-chan Envelope[T]
}

func NewPipeline[T any](ctx context.Context, source <-chan Envelope[T]) *Pipeline[T] {
	return &Pipeline[T]{
		ctx:    ctx,
		stream: source,
	}
}

func (p *Pipeline[T]) Then(operator Operator[T]) *Pipeline[T] {
	p.stream = operator(p.ctx, p.stream)
	return p
}

func (p *Pipeline[T]) UntilDone() *Pipeline[T] {
	p.stream = OrDone(p.ctx, p.stream)
	return p
}

func (p *Pipeline[T]) Merge(streams ...<-chan Envelope[T]) *Pipeline[T] {
	p.stream = FanIn(p.ctx, streams...)
	return p
}

func (p *Pipeline[T]) Flatten(sos <-chan <-chan Envelope[T]) *Pipeline[T] {
	p.stream = Bridge(p.ctx, sos)
	return p
}

func (p *Pipeline[T]) Scatter(operator Operator[T], workers int) []<-chan Envelope[T] {
	return FanOut(p.ctx, operator, p.stream, workers)
}

func (p *Pipeline[T]) Fork(streams ...<-chan Envelope[T]) (<-chan Envelope[T], <-chan Envelope[T]) {
	return Tee(p.ctx, p.stream)
}