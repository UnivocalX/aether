package universe

type Transformer[T any] func(Envelope[T]) Envelope[T]

type Predicate[T any] func(Envelope[T]) bool

type Observer[T any] func(Envelope[T])

type Reducer[T, R any] func(R, Envelope[T]) R

type Consumer[T any] func(Envelope[T]) error

// Transform value adapter
func SimpleTransformer[T any](fn func(T) T) Transformer[T] {
	return func(env Envelope[T]) Envelope[T] {
		return Envelope[T]{
			Value: fn(env.Value),
			Err:   nil,
		}
	}
}

func ValueTransformer[T any](fn func(T) (T, error)) Transformer[T] {
	return func(env Envelope[T]) Envelope[T] {
		v, e := fn(env.Value)
		return Envelope[T]{
			Value: v,
			Err:   e,
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
