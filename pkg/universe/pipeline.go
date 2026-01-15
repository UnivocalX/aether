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

func (p *Pipeline[T]) Collect(cap int) ([]Envelope[T], error) {
	return Collect(p.ctx, p.stream, cap)
}

func (p *Pipeline[T]) Drain() error {
	return Drain(p.ctx, p.stream)
}

func (p *Pipeline[T]) ForEach(fn Consume[T]) error {
	return Sink(p.ctx, fn, p.stream)
}
