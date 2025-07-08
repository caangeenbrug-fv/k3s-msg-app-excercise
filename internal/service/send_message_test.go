package service

import (
	"reflect"
	"testing"
)

func TestShiftSliceForward(t *testing.T) {
    tests := []struct {
        name     string
        slice   []int
        max_len int
        expected []int
    }{
        {"less than n", []int{1,2}, 5, []int{1,2}},
        {"more than n", []int{1,2,3,4,5,6}, 5, []int{2,3,4,5,6}},
        {"n", []int{1,2,3,4,5}, 5, []int{1,2,3,4,5}},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            result := shiftSliceForward(tc.slice, tc.max_len)
            if reflect.DeepEqual(result, tc.expected) == false {
                t.Errorf("shifting a slice forward was incorrect when given %d for max %d, which resulted in %d and not %d", tc.slice, tc.max_len, result, tc.expected)
            }
        })
    }
}
