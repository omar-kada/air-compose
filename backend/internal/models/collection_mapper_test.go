package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListMapper(t *testing.T) {
	// Test case 1: Convert a list of integers to a list of strings
	intToString := ListMapper(func(i int) string {
		return fmt.Sprintf("%d", i)
	})
	intList := []int{1, 2, 3, 4, 5}
	stringList := intToString(intList)
	assert.Equal(t, []string{"1", "2", "3", "4", "5"}, stringList)

	// Test case 2: Convert a list of strings to a list of their lengths
	stringToLength := ListMapper(func(s string) int {
		return len(s)
	})
	stringList2 := []string{"apple", "banana", "cherry"}
	lengthList := stringToLength(stringList2)
	assert.Equal(t, []int{5, 6, 6}, lengthList)
}

func TestMapMapper(t *testing.T) {
	// Test case 1: Convert a map of integers to a map of strings
	intToString := MapMapper[string](func(i int) string {
		return fmt.Sprintf("%d", i)
	})
	intMap := map[string]int{"a": 1, "b": 2, "c": 3}
	stringMap := intToString(intMap)
	assert.Equal(t, map[string]string{"a": "1", "b": "2", "c": "3"}, stringMap)

	// Test case 2: Convert a map of strings to a map of their lengths
	stringToLength := MapMapper[string](func(s string) int {
		return len(s)
	})
	stringMap2 := map[string]string{"apple": "apple", "banana": "banana", "cherry": "cherry"}
	lengthMap := stringToLength(stringMap2)
	assert.Equal(t, map[string]int{"apple": 5, "banana": 6, "cherry": 6}, lengthMap)
}
