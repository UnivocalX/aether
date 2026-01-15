package universe

func Adapter[T any](fn func(T) T) Transformer[T] {
	return func(i T) (T, error) {
		return fn(i), nil
	}
}