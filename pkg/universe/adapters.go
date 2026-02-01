package universe


func AdaptConsume[T any](c func(value T) error) Consumer[T] {
	return func(env Envelope[T]) error {
		return c(env.Value)
	}
}