package slice

import (
	"cmp"
)

// Unique preserve ordering
func Unique[T cmp.Ordered](in []T) []T {
	keys := make(map[T]struct{}, len(in))
	var list []T
	for _, entry := range in {
		if _, ok := keys[entry]; !ok {
			keys[entry] = struct{}{}
			list = append(list, entry)
		}
	}
	return list
}

// UniqueFunc preserve ordering
func UniqueFunc[T cmp.Ordered, E any](in []E, f func(E) T) []E {
	keys := make(map[T]struct{}, len(in))
	var list []E
	for _, entry := range in {
		key := f(entry)
		if _, ok := keys[key]; !ok {
			keys[key] = struct{}{}
			list = append(list, entry)
		}
	}
	return list
}
