package utils

func Map[T any, R any](a []T, t func(v T) R) []R {
	var final []R
	for _, v := range a {
		v := v
		final = append(final, t(v))
	}
	return final
}

func Array[T any](v ...T) [][]T {
	var final [][]T
	for i, val := range v {
		if i%2 != 0 {
			continue
		}
		final = append(final, []T{val, v[i+1]})
	}
	return final
}
