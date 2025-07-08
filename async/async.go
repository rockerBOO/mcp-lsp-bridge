package async

import "context"

type Result[T any] struct {
	Value T
	Error error
}

type KeyedResult[K any, T any] struct {
	Key   K
	Value T
	Error error
}

// Core async implementation - simple, no key tracking
func Map[R any](
	ctx context.Context,
	ops []func() (R, error),
) ([]Result[R], error) {
	results := make(chan Result[R], len(ops))

	for _, op := range ops {
		go func(operation func() (R, error)) {
			value, err := operation()
			results <- Result[R]{Value: value, Error: err}
		}(op)
	}

	var allResults []Result[R]
	completed := 0

	for completed < len(ops) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case result := <-results:
			allResults = append(allResults, result)
			completed++
		}
	}

	return allResults, nil
}

// Higher-level wrapper that preserves keys
func MapWithKeys[K comparable, R any](
	ctx context.Context,
	ops map[K]func() (R, error),
) ([]KeyedResult[K, R], error) {
	results := make(chan KeyedResult[K, R], len(ops))

	for key, op := range ops {
		go func(k K, operation func() (R, error)) {
			value, err := operation()
			results <- KeyedResult[K, R]{Key: k, Value: value, Error: err}
		}(key, op)
	}

	var allResults []KeyedResult[K, R]
	completed := 0

	for completed < len(ops) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case result := <-results:
			allResults = append(allResults, result)
			completed++
		}
	}

	return allResults, nil
}