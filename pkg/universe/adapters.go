package universe

// Transform value adapter
func ValueTransformer[T any](fn func(T) T) Transformer[T] {
	return func(env Envelope[T]) Envelope[T] {

		return Envelope[T]{
			Value: fn(env.Value),
			Err:   env.Err,
		}
	}
}

// Transform error adapter
func ErrorTransformer[T any](fn func(error) error) Transformer[T] {
	return func(env Envelope[T]) Envelope[T] {

		return Envelope[T]{
			Value: env.Value,
			Err:   fn(env.Err),
		}
	}
}

// Predicate Value adapter
func ValuePredicator[T any](fn func(T) bool) Predicator[T] {
	return func(env Envelope[T]) bool {
		return fn(env.Value)
	}
}

// Predicate Error adapter
func ErrorPredicator[T any](fn func(error) bool) Predicator[T] {
	return func(env Envelope[T]) bool {
		return fn(env.Err)
	}
}

// Observe value adapter
func ValueObserver[T any](fn func(T)) Observer[T] {
	return func(env Envelope[T]) {
		fn(env.Value)
	}
}

// Observe error adapter
func ErrorObserver[T any](fn func(error)) Observer[T] {
	return func(env Envelope[T]) {
		fn(env.Err)
	}
}

// consume value adapter
func ValueConsumer[T any](fn func(T) error) Consumer[T] {
	return func(env Envelope[T]) error {
		return fn(env.Value)
	}
}

// consume error adapter
func ErrorConsumer[T any](fn func(error) error) Consumer[T] {
	return func(env Envelope[T]) error {
		return fn(env.Err)
	}
}


// reduce value adapter
func ValueReducer[T, R any](fn func(R, T) R) Reducer[T, R] {
	return func(result R, env Envelope[T]) R {
		return fn(result, env.Value)
	}
}

// reduce error adapter
func ErrorReducer[T, R any](fn func(R, error) R) Reducer[T, R] {
	return func(result R, env Envelope[T]) R {
		return fn(result, env.Err)
	}
}
