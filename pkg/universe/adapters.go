package universe

type Transformer[T any] func(Envelope[T]) Envelope[T]

type Predicate[T any] func(Envelope[T]) bool

type Observer[T any] func(Envelope[T])

type Reducer[T, R any] func(R, Envelope[T]) R

type Consumer[T any] func(Envelope[T]) error


func TransformAdapter[T any](fn func(T) (T, error)) Transformer[T] {
	return func(env Envelope[T]) Envelope[T] {
		v, e := fn(env.Value)
		return Envelope[T]{
			Value: v,
			Err:   e,
		}
	}
}

// Transform value adapter
func TransformValueAdapter[T any](fn func(T) T) Transformer[T] {
	return func(env Envelope[T]) Envelope[T] {
		return Envelope[T]{
			Value: fn(env.Value),
			Err:   nil,
		}
	}
}

// Transform error adapter
func TransformErrorAdapter[T any](fn func(error) error) Transformer[T] {
	return func(env Envelope[T]) Envelope[T] {
		return Envelope[T]{
			Value: env.Value,
			Err:   fn(env.Err),
		}
	}
}

// Predicate Value adapter
func PredicateAdapter[T any](fn func(T) bool) Predicate[T] {
	return func(env Envelope[T]) bool {
		return fn(env.Value)
	}
}

// Predicate Error adapter
func PredicateErrorAdapter[T any](fn func(error) bool) Predicate[T] {
	return func(env Envelope[T]) bool {
		return fn(env.Err)
	}
}

// Observe value adapter
func ObserveAdapter[T any](fn func(T)) Observer[T] {
	return func(env Envelope[T]) {
		fn(env.Value)
	}
}

// Observe error adapter
func ObserveErrorAdapter[T any](fn func(error)) Observer[T] {
	return func(env Envelope[T]) {
		fn(env.Err)
	}
}

// consume value adapter
func ConsumeAdapter[T any](fn func(T) error) Consumer[T] {
	return func(env Envelope[T]) error {
		return fn(env.Value)
	}
}

// consume error adapter
func ConsumeErrorAdapter[T any](fn func(error) error) Consumer[T] {
	return func(env Envelope[T]) error {
		return fn(env.Err)
	}
}

// reduce value adapter
func ReduceAdapter[T, R any](fn func(R, T) R) Reducer[T, R] {
	return func(result R, env Envelope[T]) R {
		return fn(result, env.Value)
	}
}

// reduce error adapter
func ReduceErrorAdapter[T, R any](fn func(R, error) R) Reducer[T, R] {
	return func(result R, env Envelope[T]) R {
		return fn(result, env.Err)
	}
}
