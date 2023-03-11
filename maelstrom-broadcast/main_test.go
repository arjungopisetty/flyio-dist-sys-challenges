package main

import (
	"reflect"
	"testing"
)

func TestCompareSlices(t *testing.T) {
	slice1 := []int{1, 2, 3, 4, 5}
	slice2 := []int{1, 2, 3, 4}
	expectedResult := []int{5}

	diffSlice := compareSlices(slice1, slice2)
	if actual := reflect.DeepEqual(diffSlice, expectedResult); actual == false {
		t.Errorf("Expected '%v', but got '%v'", expectedResult, diffSlice)
	}
}
