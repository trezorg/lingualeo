package slice

import (
	"cmp"
)

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
