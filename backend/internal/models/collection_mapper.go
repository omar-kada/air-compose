package models

// ListMapper creates a function that applies a given transformation function to each element of a slice.
//
// Parameters:
//   - fn: A function that takes an element of type T and returns an element of type V.
//
// Returns:
//   - A function that takes a slice of type []T and returns a slice of type []V.
func ListMapper[T any, V any](fn func(T) V) func([]T) []V {
	return func(inputs []T) []V {
		output := make([]V, 0)
		for _, input := range inputs {
			output = append(output, fn(input))
		}
		return output
	}
}

// MapMapper creates a function that applies a given transformation function to each value of a map.
//
// Parameters:
//   - fn: A function that takes a value of type T and returns a value of type V.
//
// Returns:
//   - A function that takes a map of type map[K]T and returns a map of type map[K]V.
func MapMapper[K comparable, T any, V any](fn func(T) V) func(map[K]T) map[K]V {
	return func(inputMap map[K]T) map[K]V {
		resMap := make(map[K]V)
		for key, value := range inputMap {
			resMap[key] = fn(value)
		}
		return resMap
	}
}
