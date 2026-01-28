package utils

func Filter[S []T, T any](s S, filter func(T) bool) S {
	newS := make(S, 0, len(s))
	for _, v := range s {
		if filter(v) {
			newS = append(newS, v)
		}
	}
	return newS
}

func Map[S []T1, T1, T2 any](s S, mapFn func(T1) T2) []T2 {
	newS := make([]T2, 0, len(s))
	for _, v := range s {
		newS = append(newS, mapFn(v))
	}
	return newS
}
