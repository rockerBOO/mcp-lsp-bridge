package utils

import (
	"fmt"
	"rockerboo/mcp-lsp-bridge/async"
)

type FlattenedResult[T any] struct {
	Values []T
	Errors []error
}

func FlattenResults[T any](results []async.Result[[]T]) FlattenedResult[T] {
	var flattened []T
	var errors []error

	for _, result := range results {
		if result.Error == nil {
			flattened = append(flattened, result.Value...)
		} else {
			errors = append(errors, result.Error)
		}
	}

	return FlattenedResult[T]{Values: flattened, Errors: errors}
}

func FlattenKeyedResults[K any, T any](results []async.KeyedResult[K, []T]) FlattenedResult[T] {
	var flattened []T
	var errors []error

	for _, result := range results {
		if result.Error == nil {
			flattened = append(flattened, result.Value...)
		} else {
			errors = append(errors, fmt.Errorf("key %v: %w", result.Key, result.Error))
		}
	}

	return FlattenedResult[T]{Values: flattened, Errors: errors}
}
