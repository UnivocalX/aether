package universe

func TransformValue[T, U any](fn func(value T) (U, error)) Transformer[T, U] {
	return func(meta *Meta, env Envelope[T]) Envelope[U] {
		v, e := fn(env.Value)
		return Envelope[U]{
			Value: v,
			Err: e,
		}
	}
}

func FilterValue[T any](fn func(value T) bool) Predicate[T] {
	return func(meta *Meta, env Envelope[T]) bool {
		return fn(env.Value)
	}
}

func ObserveValue[T any](fn func(value T)) Observer[T] {
	return func(meta *Meta, env Envelope[T]) {
		fn(env.Value)
	}
}

func ConsumeValue[T any](c func(value T) error) Consumer[T] {
	return func(env Envelope[T]) error {
		return c(env.Value)
	}
}
