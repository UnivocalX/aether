package universe

// Transform value adapter
func ValueTransform[T any](fn func(T) T) Transform[T] {
	return func(env Envelope[T]) Envelope[T] {

		return Envelope[T]{
			Value: fn(env.Value),
			Err:   env.Err,
		}
	}
}

// Transform error adapter
func ErrorTransform[T any](fn func(error) error) Transform[T] {
	return func(env Envelope[T]) Envelope[T] {

		return Envelope[T]{
			Value: env.Value,
			Err:   fn(env.Err),
		}
	}
}

// Predicate Value adapter
func ValuePredicate[T any](fn func(T) bool) Predicate[T] {
	return func(env Envelope[T]) bool {
		return fn(env.Value)
	}
}

// Predicate Error adapter
func ErrorPredicate[T any](fn func(error) bool) Predicate[T] {
	return func(env Envelope[T]) bool {
		return fn(env.Err)
	}
}

// Observe value adapter
func ValueObserve[T any](fn func(T)) Observe[T] {
	return func(env Envelope[T]) {
		fn(env.Value)
	}
}

// Observe error adapter
func ErrorObserve[T any](fn func(error)) Observe[T] {
	return func(env Envelope[T]) {
		fn(env.Err)
	}
}

// consume value adapter
func ValueConsume[T any](fn func(T) error) Consume[T] {
	return func(env Envelope[T]) error {
		return fn(env.Value)
	}
}

// consume error adapter
func ErrorConsumer[T any](fn func(error) error) Consume[T] {
	return func(env Envelope[T]) error {
		return fn(env.Err)
	}
}