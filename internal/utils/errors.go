package utils

import "errors"

func MapErr[T error](err error) (T, bool) {
	var target T
	ok := errors.As(err, &target)
	return target, ok
}
