package universe

import (
	"context"
	"sync"
)

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

func Guard[T any](ctx context.Context, stream <-chan Envelope[T]) <-chan Envelope[T] {
	out := make(chan Envelope[T])

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case env, ok := <-stream:
				if !ok {
					return
				}
				select {
				case out <- env:
				case <-ctx.Done():
				}
			}
		}
	}()

	return out
}

func FanIn[T any](
	ctx context.Context,
	streams ...<-chan Envelope[T],
) <-chan Envelope[T] {

	var wg sync.WaitGroup
	out := make(chan Envelope[T])

	wg.Add(len(streams))
	for _, s := range streams {
		go func(c <-chan Envelope[T]) {
			defer wg.Done()
			for env := range c {
				select {
				case <-ctx.Done():
					return
				case out <- env:
				}
			}
		}(s)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func FanOut[T any](
	ctx context.Context,
	fn func(context.Context, <-chan Envelope[T]) <-chan Envelope[T],
	stream <-chan Envelope[T],
	workers int,
) []<-chan Envelope[T] {

	out := make([]<-chan Envelope[T], workers)

	for i := 0; i < workers; i++ {
		out[i] = fn(ctx, stream)
	}

	return out
}

func Tee[T any](
	ctx context.Context,
	stream <-chan Envelope[T],
) (<-chan Envelope[T], <-chan Envelope[T]) {

	out1 := make(chan Envelope[T])
	out2 := make(chan Envelope[T])

	go func() {
		defer close(out1)
		defer close(out2)

		for env := range Guard(ctx, stream) {
			var o1, o2 = out1, out2
			for i := 2; i > 0; i-- {
				select {
				case <-ctx.Done():
					return
				case o1 <- env:
					o1 = nil
				case o2 <- env:
					o2 = nil
				}
			}
		}
	}()

	return out1, out2
}

func Bridge[T any](
	ctx context.Context,
	sos <-chan <-chan Envelope[T],
) <-chan Envelope[T] {

	out := make(chan Envelope[T])

	go func() {
		defer close(out)

		for {
			var stream <-chan Envelope[T]

			select {
			case <-ctx.Done():
				return
			case s, ok := <-sos:
				if !ok {
					return
				}
				stream = s
			}

			for env := range Guard(ctx, stream) {
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