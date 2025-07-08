package util

import (
	"fmt"
)

func IndexOf[T comparable](collection []T, value T) (int, error) {
	for i, v := range collection {
		if v == value {
			return i, nil
		}
	}

	return -1, fmt.Errorf("failed to find index of value %v in collection %v", value, collection)
}

func Filter[T comparable](collection []T, value T) []T {
	var result = []T{}
	for _, v := range collection {
		if v != value {
			result = append(result, v)
		}
	}

	return result
}
