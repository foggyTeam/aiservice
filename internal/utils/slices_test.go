package utils

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter_WithIntegers(t *testing.T) {
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result := Filter(input, func(x int) bool {
		return x%2 == 0 // filter even numbers
	})

	expected := []int{2, 4, 6, 8, 10}
	assert.Equal(t, expected, result, "Filter should return only even numbers")
}

func TestFilter_WithStrings(t *testing.T) {
	input := []string{"apple", "apricot", "banana", "blueberry", "cherry"}
	result := Filter(input, func(s string) bool {
		return strings.HasPrefix(s, "a") // filter strings starting with 'a'
	})

	expected := []string{"apple", "apricot"}
	assert.Equal(t, expected, result, "Filter should return only strings starting with 'a'")
}

func TestFilter_EmptySlice(t *testing.T) {
	input := []int{}
	result := Filter(input, func(x int) bool {
		return x > 5
	})

	assert.Equal(t, []int{}, result, "Filter should return empty slice for empty input")
	assert.Equal(t, 0, len(result), "Result should have length 0")
}

func TestFilter_NoMatches(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	result := Filter(input, func(x int) bool {
		return x > 100 // no elements will match
	})

	assert.Equal(t, []int{}, result, "Filter should return empty slice when no elements match")
	assert.Equal(t, 0, len(result), "Result should have length 0")
}

func TestFilter_AllMatch(t *testing.T) {
	input := []int{2, 4, 6, 8, 10}
	result := Filter(input, func(x int) bool {
		return x > 0 // all elements will match
	})

	assert.Equal(t, input, result, "Filter should return all elements when all match")
	assert.Equal(t, len(input), len(result), "Result should have same length as input")
}

func TestFilter_SingleElement(t *testing.T) {
	input := []int{5}
	result := Filter(input, func(x int) bool {
		return x > 3
	})

	assert.Equal(t, []int{5}, result, "Filter should return single element if it matches")

	result2 := Filter(input, func(x int) bool {
		return x > 10
	})

	assert.Equal(t, []int{}, result2, "Filter should return empty slice if single element doesn't match")
}

func TestFilter_WithComplexObjects(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	input := []Person{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
		{Name: "Charlie", Age: 35},
		{Name: "Diana", Age: 28},
	}

	result := Filter(input, func(p Person) bool {
		return p.Age >= 30
	})

	expected := []Person{
		{Name: "Alice", Age: 30},
		{Name: "Charlie", Age: 35},
	}
	assert.Equal(t, expected, result, "Filter should correctly filter complex objects")
}

func TestFilter_PreservesOrder(t *testing.T) {
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	result := Filter(input, func(x int) bool {
		return x%2 == 0 // even numbers
	})

	assert.Equal(t, []int{2, 4, 6, 8}, result, "Filter should preserve order of matched elements")
}

func TestFilter_WithBoolean(t *testing.T) {
	input := []bool{true, false, true, true, false, false, true}
	result := Filter(input, func(b bool) bool {
		return b
	})

	expected := []bool{true, true, true, true}
	assert.Equal(t, expected, result, "Filter should correctly filter boolean values")
}

func TestFilter_WithNegativePredicate(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	result := Filter(input, func(x int) bool {
		return x <= 2
	})

	expected := []int{1, 2}
	assert.Equal(t, expected, result, "Filter should work with negative conditions")
}

func TestFilter_DoesNotModifyOriginal(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	inputCopy := make([]int, len(input))
	copy(inputCopy, input)

	Filter(input, func(x int) bool {
		return x > 2
	})

	assert.Equal(t, inputCopy, input, "Filter should not modify the original slice")
}

func TestMap_WithIntegers(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	result := Map(input, func(x int) int {
		return x * 2 // double each number
	})

	expected := []int{2, 4, 6, 8, 10}
	assert.Equal(t, expected, result, "Map should double all integers")
}

func TestMap_WithStringTransformation(t *testing.T) {
	input := []string{"hello", "world", "go"}
	result := Map(input, func(s string) string {
		return strings.ToUpper(s) // convert to uppercase
	})

	expected := []string{"HELLO", "WORLD", "GO"}
	assert.Equal(t, expected, result, "Map should convert strings to uppercase")
}

func TestMap_IntToString(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	result := Map(input, func(x int) string {
		return strconv.Itoa(x)
	})

	expected := []string{"1", "2", "3", "4", "5"}
	assert.Equal(t, expected, result, "Map should convert integers to strings")
}

func TestMap_StringToInt(t *testing.T) {
	input := []string{"1", "2", "3", "4", "5"}
	result := Map(input, func(s string) int {
		val, _ := strconv.Atoi(s)
		return val
	})

	expected := []int{1, 2, 3, 4, 5}
	assert.Equal(t, expected, result, "Map should convert strings to integers")
}

func TestMap_EmptySlice(t *testing.T) {
	input := []int{}
	result := Map(input, func(x int) int {
		return x * 2
	})

	assert.Equal(t, []int{}, result, "Map should return empty slice for empty input")
	assert.Equal(t, 0, len(result), "Result should have length 0")
}

func TestMap_SingleElement(t *testing.T) {
	input := []int{42}
	result := Map(input, func(x int) int {
		return x * 2
	})

	assert.Equal(t, []int{84}, result, "Map should correctly transform single element")
}

func TestMap_WithComplexObjects(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	input := []Person{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
		{Name: "Charlie", Age: 35},
	}

	result := Map(input, func(p Person) string {
		return p.Name
	})

	expected := []string{"Alice", "Bob", "Charlie"}
	assert.Equal(t, expected, result, "Map should extract names from persons")
}

func TestMap_WithStructExtraction(t *testing.T) {
	type Item struct {
		ID    int
		Price float64
	}

	input := []Item{
		{ID: 1, Price: 10.5},
		{ID: 2, Price: 20.0},
		{ID: 3, Price: 15.75},
	}

	result := Map(input, func(item Item) float64 {
		return item.Price * 1.1 // apply 10% markup
	})

	// Use InDelta for float comparisons due to floating-point precision
	expected := []float64{11.55, 22.0, 17.325}
	assert.Equal(t, len(expected), len(result), "Map should return correct number of items")
	for i := range expected {
		assert.InDelta(t, expected[i], result[i], 0.0001, "Map should apply markup to prices correctly")
	}
}

func TestMap_PreservesOrder(t *testing.T) {
	input := []int{5, 1, 3, 9, 2}
	result := Map(input, func(x int) int {
		return x * 10
	})

	expected := []int{50, 10, 30, 90, 20}
	assert.Equal(t, expected, result, "Map should preserve order of elements")
}

func TestMap_ToBoolean(t *testing.T) {
	input := []int{0, 1, 2, 0, 5}
	result := Map(input, func(x int) bool {
		return x > 0
	})

	expected := []bool{false, true, true, false, true}
	assert.Equal(t, expected, result, "Map should convert integers to booleans correctly")
}

func TestMap_DoesNotModifyOriginal(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	inputCopy := make([]int, len(input))
	copy(inputCopy, input)

	Map(input, func(x int) int {
		return x * 10
	})

	assert.Equal(t, inputCopy, input, "Map should not modify the original slice")
}

func TestMap_WithFloatRounding(t *testing.T) {
	input := []float64{1.4, 1.5, 2.6, 3.1}
	result := Map(input, func(f float64) int {
		return int(f + 0.5) // simple rounding
	})

	expected := []int{1, 2, 3, 3}
	assert.Equal(t, expected, result, "Map should round floats to integers")
}

func TestFilterAndMapCombined(t *testing.T) {
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// First filter even numbers, then double them
	filtered := Filter(input, func(x int) bool {
		return x%2 == 0
	})
	result := Map(filtered, func(x int) int {
		return x * 2
	})

	expected := []int{4, 8, 12, 16, 20}
	assert.Equal(t, expected, result, "Combined Filter and Map should work correctly")
}

func TestMap_LargeSlice(t *testing.T) {
	input := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		input[i] = i
	}

	result := Map(input, func(x int) int {
		return x * 2
	})

	assert.Equal(t, 1000, len(result), "Map should handle large slices")
	assert.Equal(t, 0, result[0], "First element should be 0")
	assert.Equal(t, 1998, result[999], "Last element should be 1998")
}

func TestFilter_LargeSlice(t *testing.T) {
	input := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		input[i] = i
	}

	result := Filter(input, func(x int) bool {
		return x%2 == 0 // even numbers
	})

	assert.Equal(t, 500, len(result), "Filter should return 500 even numbers")
	assert.Equal(t, 0, result[0], "First element should be 0")
	assert.Equal(t, 998, result[499], "Last element should be 998")
}
