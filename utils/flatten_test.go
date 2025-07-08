package utils

import (
	"errors"
	"testing"

	"rockerboo/mcp-lsp-bridge/async"
)

func TestFlattenResults(t *testing.T) {
	t.Run("all successful results", func(t *testing.T) {
		input := []async.Result[[]int]{
			{Value: []int{1, 2, 3}, Error: nil},
			{Value: []int{4, 5}, Error: nil},
			{Value: []int{6}, Error: nil},
		}

		result := FlattenResults(input)

		expectedValues := []int{1, 2, 3, 4, 5, 6}
		if len(result.Values) != len(expectedValues) {
			t.Fatalf("expected %d values, got %d", len(expectedValues), len(result.Values))
		}

		for i, expected := range expectedValues {
			if result.Values[i] != expected {
				t.Errorf("index %d: expected %d, got %d", i, expected, result.Values[i])
			}
		}

		if len(result.Errors) != 0 {
			t.Errorf("expected no errors, got %d", len(result.Errors))
		}
	})

	t.Run("mixed success and failure", func(t *testing.T) {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")

		input := []async.Result[[]string]{
			{Value: []string{"a", "b"}, Error: nil},
			{Value: nil, Error: err1},
			{Value: []string{"c"}, Error: nil},
			{Value: nil, Error: err2},
		}

		result := FlattenResults(input)

		expectedValues := []string{"a", "b", "c"}
		if len(result.Values) != len(expectedValues) {
			t.Fatalf("expected %d values, got %d", len(expectedValues), len(result.Values))
		}

		for i, expected := range expectedValues {
			if result.Values[i] != expected {
				t.Errorf("index %d: expected %s, got %s", i, expected, result.Values[i])
			}
		}

		if len(result.Errors) != 2 {
			t.Fatalf("expected 2 errors, got %d", len(result.Errors))
		}

		if result.Errors[0] != err1 {
			t.Errorf("first error: expected %v, got %v", err1, result.Errors[0])
		}
		if result.Errors[1] != err2 {
			t.Errorf("second error: expected %v, got %v", err2, result.Errors[1])
		}
	})

	t.Run("empty input", func(t *testing.T) {
		input := []async.Result[[]int]{}

		result := FlattenResults(input)

		if len(result.Values) != 0 {
			t.Errorf("expected empty values, got %d", len(result.Values))
		}

		if len(result.Errors) != 0 {
			t.Errorf("expected no errors, got %d", len(result.Errors))
		}
	})

	t.Run("empty slices", func(t *testing.T) {
		input := []async.Result[[]int]{
			{Value: []int{}, Error: nil},
			{Value: []int{}, Error: nil},
		}

		result := FlattenResults(input)

		if len(result.Values) != 0 {
			t.Errorf("expected empty values, got %d", len(result.Values))
		}

		if len(result.Errors) != 0 {
			t.Errorf("expected no errors, got %d", len(result.Errors))
		}
	})

	t.Run("all failures", func(t *testing.T) {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")

		input := []async.Result[[]int]{
			{Value: nil, Error: err1},
			{Value: nil, Error: err2},
		}

		result := FlattenResults(input)

		if len(result.Values) != 0 {
			t.Errorf("expected empty values, got %d", len(result.Values))
		}

		if len(result.Errors) != 2 {
			t.Fatalf("expected 2 errors, got %d", len(result.Errors))
		}

		if result.Errors[0] != err1 {
			t.Errorf("first error: expected %v, got %v", err1, result.Errors[0])
		}
		if result.Errors[1] != err2 {
			t.Errorf("second error: expected %v, got %v", err2, result.Errors[1])
		}
	})
}

func TestFlattenKeyedResults(t *testing.T) {
	t.Run("all successful results", func(t *testing.T) {
		input := []async.KeyedResult[string, []int]{
			{Key: "first", Value: []int{1, 2, 3}, Error: nil},
			{Key: "second", Value: []int{4, 5}, Error: nil},
			{Key: "third", Value: []int{6}, Error: nil},
		}

		result := FlattenKeyedResults(input)

		expectedValues := []int{1, 2, 3, 4, 5, 6}
		if len(result.Values) != len(expectedValues) {
			t.Fatalf("expected %d values, got %d", len(expectedValues), len(result.Values))
		}

		for i, expected := range expectedValues {
			if result.Values[i] != expected {
				t.Errorf("index %d: expected %d, got %d", i, expected, result.Values[i])
			}
		}

		if len(result.Errors) != 0 {
			t.Errorf("expected no errors, got %d", len(result.Errors))
		}
	})

	t.Run("mixed success and failure with key context", func(t *testing.T) {
		err1 := errors.New("connection failed")
		err2 := errors.New("timeout")

		input := []async.KeyedResult[string, []string]{
			{Key: "go", Value: []string{"func1", "func2"}, Error: nil},
			{Key: "python", Value: nil, Error: err1},
			{Key: "javascript", Value: []string{"func3"}, Error: nil},
			{Key: "rust", Value: nil, Error: err2},
		}

		result := FlattenKeyedResults(input)

		expectedValues := []string{"func1", "func2", "func3"}
		if len(result.Values) != len(expectedValues) {
			t.Fatalf("expected %d values, got %d", len(expectedValues), len(result.Values))
		}

		for i, expected := range expectedValues {
			if result.Values[i] != expected {
				t.Errorf("index %d: expected %s, got %s", i, expected, result.Values[i])
			}
		}

		if len(result.Errors) != 2 {
			t.Fatalf("expected 2 errors, got %d", len(result.Errors))
		}

		// Check that errors contain key context
		error1Str := result.Errors[0].Error()
		error2Str := result.Errors[1].Error()

		if !contains(error1Str, "python") || !contains(error1Str, "connection failed") {
			t.Errorf("first error should contain key context: %s", error1Str)
		}
		if !contains(error2Str, "rust") || !contains(error2Str, "timeout") {
			t.Errorf("second error should contain key context: %s", error2Str)
		}
	})

	t.Run("different key types", func(t *testing.T) {
		input := []async.KeyedResult[int, []string]{
			{Key: 1, Value: []string{"one"}, Error: nil},
			{Key: 2, Value: []string{"two"}, Error: nil},
		}

		result := FlattenKeyedResults(input)

		expectedValues := []string{"one", "two"}
		if len(result.Values) != len(expectedValues) {
			t.Fatalf("expected %d values, got %d", len(expectedValues), len(result.Values))
		}

		for i, expected := range expectedValues {
			if result.Values[i] != expected {
				t.Errorf("index %d: expected %s, got %s", i, expected, result.Values[i])
			}
		}
	})

	t.Run("empty input", func(t *testing.T) {
		input := []async.KeyedResult[string, []int]{}

		result := FlattenKeyedResults(input)

		if len(result.Values) != 0 {
			t.Errorf("expected empty values, got %d", len(result.Values))
		}

		if len(result.Errors) != 0 {
			t.Errorf("expected no errors, got %d", len(result.Errors))
		}
	})

	t.Run("complex value types", func(t *testing.T) {
		type Symbol struct {
			Name string
			Type string
		}

		input := []async.KeyedResult[string, []Symbol]{
			{
				Key: "go",
				Value: []Symbol{
					{Name: "main", Type: "function"},
					{Name: "Config", Type: "struct"},
				},
				Error: nil,
			},
			{
				Key: "python",
				Value: []Symbol{
					{Name: "parse", Type: "function"},
				},
				Error: nil,
			},
		}

		result := FlattenKeyedResults(input)

		expectedCount := 3
		if len(result.Values) != expectedCount {
			t.Fatalf("expected %d values, got %d", expectedCount, len(result.Values))
		}

		// Check first symbol
		if result.Values[0].Name != "main" || result.Values[0].Type != "function" {
			t.Errorf("first symbol incorrect: got %+v", result.Values[0])
		}

		// Check second symbol
		if result.Values[1].Name != "Config" || result.Values[1].Type != "struct" {
			t.Errorf("second symbol incorrect: got %+v", result.Values[1])
		}

		// Check third symbol
		if result.Values[2].Name != "parse" || result.Values[2].Type != "function" {
			t.Errorf("third symbol incorrect: got %+v", result.Values[2])
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 indexOf(s, substr) != -1)))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func BenchmarkFlattenResults(b *testing.B) {
	// Create a large slice of results
	input := make([]async.Result[[]int], 1000)
	for i := 0; i < 1000; i++ {
		input[i] = async.Result[[]int]{
			Value: []int{i, i + 1, i + 2},
			Error: nil,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FlattenResults(input)
	}
}

func BenchmarkFlattenKeyedResults(b *testing.B) {
	// Create a large slice of keyed results
	input := make([]async.KeyedResult[string, []int], 1000)
	for i := 0; i < 1000; i++ {
		input[i] = async.KeyedResult[string, []int]{
			Key:   "key" + string(rune(i)),
			Value: []int{i, i + 1, i + 2},
			Error: nil,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FlattenKeyedResults(input)
	}
}