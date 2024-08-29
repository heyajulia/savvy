package fp

func Map[T1, T2 any](f func(T1) T2, values []T1) []T2 {
	mapped := make([]T2, 0, len(values))

	for _, value := range values {
		mapped = append(mapped, f(value))
	}

	return mapped
}
