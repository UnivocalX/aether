package universe

import "context"

// Adapter adapts a pure function into a stage.
func Adapter[T any](fn func(T) T) StageFn[T] {
	return func(ctx context.Context, in T) Envelope[T] {
		return Envelope[T]{Value: fn(in)}
	}
}

// ErrorAdapter adapts an error-returning function (no context) into a stage.
func ErrorAdapter[T any](fn func(T) (T, error)) StageFn[T] {
	return func(ctx context.Context, in T) Envelope[T] {
		v, err := fn(in)
		return Envelope[T]{Value: v, Err: err}
	}
}

// ContextErrorAdapter adapts a context-aware error-returning function into a stage.
func ContextErrorAdapter[T any](fn func(context.Context, T) (T, error)) StageFn[T] {
	return func(ctx context.Context, in T) Envelope[T] {
		v, err := fn(ctx, in)
		return Envelope[T]{Value: v, Err: err}
	}
}

// SideEffectAdapter adapts a side-effect-only function into a stage.
func SideEffectAdapter[T any](fn func(context.Context, T) error) StageFn[T] {
	return func(ctx context.Context, in T) Envelope[T] {
		return Envelope[T]{Value: in, Err: fn(ctx, in)}
	}
}