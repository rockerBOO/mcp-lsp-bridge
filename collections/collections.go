package collections

// TransformMap transforms each value in a map using the provided operation
func TransformMap[K comparable, V any, F any](
	items map[K]V,
	operation func(V) F,
) map[K]F {
	result := make(map[K]F)
	for key, value := range items {
		result[key] = operation(value)
	}
	return result
}

// ToString converts any slice to a slice of strings using string conversion
// Works best with types that can be converted to string (e.g., custom string types)
func ToString[T ~string](items []T) []string {
	result := make([]string, len(items))
	for i, item := range items {
		result[i] = string(item)
	}
	return result
}