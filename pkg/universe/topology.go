package universe

import (
	"context"
	"sync"
)

func Repeat[T any](ctx context.Context, fn func() T, times uint) <-chan T {
	out := make(chan T)

	go func() {
		defer close(out)
		for i := 0; i < int(times); i++{
			select {
			case <-ctx.Done():
				return
			case out <- fn():
			}
		}
	}()

	return out
}

func OrDone[T any](ctx context.Context, stream <-chan T) <-chan T {
	out := make(chan T)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case value, ok := <-stream:
				if !ok {
					return
				}
				select {
				case out <- value:
				case <-ctx.Done():
				}
			}
		}
	}()

	return out
}

func FanIn[T any](
	ctx context.Context,
	streams ...<-chan T,
) <-chan T {

	var wg sync.WaitGroup
	out := make(chan T)

	wg.Add(len(streams))
	for _, s := range streams {

		go func(c <-chan T) {
			defer wg.Done()
			for value := range c {
				select {
				case <-ctx.Done():
					return
				case out <- value:
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
	fn func(context.Context, <-chan T) <-chan T,
	stream <-chan T,
	workers int,
) []<-chan T {

	out := make([]<-chan T, workers)

	for i := 0; i < workers; i++ {
		out[i] = fn(ctx, stream)
	}

	return out
}

func Tee[T any](
	ctx context.Context,
	stream <-chan T,
) (<-chan T, <-chan T) {

	out1 := make(chan T)
	out2 := make(chan T)

	go func() {
		defer close(out1)
		defer close(out2)

		for value := range OrDone(ctx, stream) {
			var o1, o2 = out1, out2
			for i := 2; i > 0; i-- {
				select {
				case <-ctx.Done():
					return
				case o1 <- value:
					o1 = nil
				case o2 <- value:
					o2 = nil
				}
			}
		}
	}()

	return out1, out2
}

func Bridge[T any](
	ctx context.Context,
	sos <-chan <-chan T,
) <-chan T {

	out := make(chan T)

	go func() {
		defer close(out)

		for {
			var stream <-chan T

			select {
			case <-ctx.Done():
				return
			case s, ok := <-sos:
				if !ok {
					return
				}
				stream = s
			}

			for value := range OrDone(ctx, stream) {
				select {
				case <-ctx.Done():
					return
				case out <- value:
				}
			}
		}
	}()

	return out
}