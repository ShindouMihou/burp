package utils

import "strings"

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

func AllMatches[T any](a []T, predicate func(b T) bool) []T {
	var l = []T{}
	for _, v := range a {
		v := v
		if predicate(v) {
			l = append(l, v)
		}
	}
	return l
}

func Filter[T any](a []T, predicate func(b T) bool) *T {
	for _, v := range a {
		v := v
		if predicate(v) {
			return &v
		}
	}
	return nil
}

func AnyMatch[T any](a []T, predicate func(b T) bool) bool {
	match := Filter(a, predicate)
	return match != nil
}

func AnyMatchString(a []string, match string) bool {
	return AnyMatch(a, func(b string) bool {
		return b == match
	})
}

func AnyMatchStringCaseInsensitive(a []string, match string) bool {
	return AnyMatch(a, func(b string) bool {
		return strings.EqualFold(b, match)
	})
}
