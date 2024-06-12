package fp

func Map[T1, T2 any](f func(T1) T2, values []T1) []T2 {
	mapped := make([]T2, 0, len(values))

	for _, value := range values {
		mapped = append(mapped, f(value))
	}

	return mapped
}

func Reduce[T1, T2 any](f func(T1, T2) T1, initial T1, values []T2) T1 {
	acc := initial

	for _, value := range values {
		acc = f(acc, value)
	}

	return acc
}

func Where[T any](f func(T) bool, values []T) []T {
	var filtered []T

	for _, value := range values {
		if f(value) {
			filtered = append(filtered, value)
		}
	}

	return filtered
}
