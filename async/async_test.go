package async

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMap(t *testing.T) {
	t.Run("successful operations", func(t *testing.T) {
		ops := []func() (int, error){
			func() (int, error) { return 1, nil },
			func() (int, error) { return 2, nil },
			func() (int, error) { return 3, nil },
		}

		ctx := context.Background()
		results, err := Map(ctx, ops)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 3 {
			t.Fatalf("expected 3 results, got %d", len(results))
		}

		// Check that all values are present (order may vary due to concurrency)
		values := make(map[int]bool)
		for _, result := range results {
			if result.Error != nil {
				t.Errorf("unexpected error in result: %v", result.Error)
			}
			values[result.Value] = true
		}

		for i := 1; i <= 3; i++ {
			if !values[i] {
				t.Errorf("missing value %d", i)
			}
		}
	})

	t.Run("mixed success and failure", func(t *testing.T) {
		testErr := errors.New("test error")
		ops := []func() (int, error){
			func() (int, error) { return 1, nil },
			func() (int, error) { return 0, testErr },
			func() (int, error) { return 3, nil },
		}

		ctx := context.Background()
		results, err := Map(ctx, ops)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 3 {
			t.Fatalf("expected 3 results, got %d", len(results))
		}

		successCount := 0
		errorCount := 0
		for _, result := range results {
			if result.Error != nil {
				errorCount++
				if result.Error != testErr {
					t.Errorf("expected test error, got %v", result.Error)
				}
			} else {
				successCount++
			}
		}

		if successCount != 2 {
			t.Errorf("expected 2 successful operations, got %d", successCount)
		}
		if errorCount != 1 {
			t.Errorf("expected 1 failed operation, got %d", errorCount)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ops := []func() (int, error){
			func() (int, error) {
				time.Sleep(100 * time.Millisecond)
				return 1, nil
			},
			func() (int, error) {
				time.Sleep(100 * time.Millisecond)
				return 2, nil
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		results, err := Map(ctx, ops)

		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}

		if results != nil {
			t.Errorf("expected nil results on cancellation, got %v", results)
		}
	})

	t.Run("empty operations", func(t *testing.T) {
		ops := []func() (int, error){}

		ctx := context.Background()
		results, err := Map(ctx, ops)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("expected empty results, got %d", len(results))
		}
	})
}

func TestMapWithKeys(t *testing.T) {
	t.Run("successful operations", func(t *testing.T) {
		ops := map[string]func() (int, error){
			"first":  func() (int, error) { return 1, nil },
			"second": func() (int, error) { return 2, nil },
			"third":  func() (int, error) { return 3, nil },
		}

		ctx := context.Background()
		results, err := MapWithKeys(ctx, ops)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 3 {
			t.Fatalf("expected 3 results, got %d", len(results))
		}

		// Check that all keys and values are present
		resultMap := make(map[string]int)
		for _, result := range results {
			if result.Error != nil {
				t.Errorf("unexpected error in result for key %s: %v", result.Key, result.Error)
			}
			resultMap[result.Key] = result.Value
		}

		expectedMap := map[string]int{
			"first":  1,
			"second": 2,
			"third":  3,
		}

		for key, expectedValue := range expectedMap {
			if actualValue, exists := resultMap[key]; !exists {
				t.Errorf("missing key %s", key)
			} else if actualValue != expectedValue {
				t.Errorf("key %s: expected %d, got %d", key, expectedValue, actualValue)
			}
		}
	})

	t.Run("mixed success and failure", func(t *testing.T) {
		testErr := errors.New("test error")
		ops := map[string]func() (int, error){
			"success": func() (int, error) { return 1, nil },
			"failure": func() (int, error) { return 0, testErr },
		}

		ctx := context.Background()
		results, err := MapWithKeys(ctx, ops)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}

		resultMap := make(map[string]KeyedResult[string, int])
		for _, result := range results {
			resultMap[result.Key] = result
		}

		// Check success case
		if successResult, exists := resultMap["success"]; !exists {
			t.Error("missing success result")
		} else {
			if successResult.Error != nil {
				t.Errorf("unexpected error in success result: %v", successResult.Error)
			}
			if successResult.Value != 1 {
				t.Errorf("expected success value 1, got %d", successResult.Value)
			}
		}

		// Check failure case
		if failureResult, exists := resultMap["failure"]; !exists {
			t.Error("missing failure result")
		} else {
			if failureResult.Error != testErr {
				t.Errorf("expected test error, got %v", failureResult.Error)
			}
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ops := map[string]func() (int, error){
			"slow": func() (int, error) {
				time.Sleep(100 * time.Millisecond)
				return 1, nil
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		results, err := MapWithKeys(ctx, ops)

		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}

		if results != nil {
			t.Errorf("expected nil results on cancellation, got %v", results)
		}
	})

	t.Run("empty operations", func(t *testing.T) {
		ops := map[string]func() (int, error){}

		ctx := context.Background()
		results, err := MapWithKeys(ctx, ops)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("expected empty results, got %d", len(results))
		}
	})
}

func TestConcurrency(t *testing.T) {
	t.Run("operations run concurrently", func(t *testing.T) {
		start := time.Now()
		
		ops := []func() (int, error){
			func() (int, error) {
				time.Sleep(50 * time.Millisecond)
				return 1, nil
			},
			func() (int, error) {
				time.Sleep(50 * time.Millisecond)
				return 2, nil
			},
			func() (int, error) {
				time.Sleep(50 * time.Millisecond)
				return 3, nil
			},
		}

		ctx := context.Background()
		results, err := Map(ctx, ops)

		elapsed := time.Since(start)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 3 {
			t.Fatalf("expected 3 results, got %d", len(results))
		}

		// If operations ran sequentially, it would take ~150ms
		// If concurrent, it should take ~50ms (plus some overhead)
		if elapsed > 100*time.Millisecond {
			t.Errorf("operations appear to be running sequentially, took %v", elapsed)
		}
	})
}