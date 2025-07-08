package collections

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
)

func TestTransformMap(t *testing.T) {
	t.Run("transform string to int", func(t *testing.T) {
		input := map[string]string{
			"one":   "1",
			"two":   "2",
			"three": "3",
		}

		result := TransformMap(input, func(s string) int {
			switch s {
			case "1":
				return 1
			case "2":
				return 2
			case "3":
				return 3
			default:
				return 0
			}
		})

		expected := map[string]int{
			"one":   1,
			"two":   2,
			"three": 3,
		}

		if len(result) != len(expected) {
			t.Fatalf("expected %d results, got %d", len(expected), len(result))
		}

		for key, expectedValue := range expected {
			if actualValue, exists := result[key]; !exists {
				t.Errorf("missing key %s", key)
			} else if actualValue != expectedValue {
				t.Errorf("key %s: expected %d, got %d", key, expectedValue, actualValue)
			}
		}
	})

	t.Run("transform to function", func(t *testing.T) {
		input := map[string]int{
			"double": 2,
			"triple": 3,
			"quad":   4,
		}

		result := TransformMap(input, func(multiplier int) func(int) int {
			return func(x int) int {
				return x * multiplier
			}
		})

		// Test the generated functions
		if result["double"](5) != 10 {
			t.Errorf("double function failed: expected 10, got %d", result["double"](5))
		}
		if result["triple"](5) != 15 {
			t.Errorf("triple function failed: expected 15, got %d", result["triple"](5))
		}
		if result["quad"](5) != 20 {
			t.Errorf("quad function failed: expected 20, got %d", result["quad"](5))
		}
	})

	t.Run("transform to error function", func(t *testing.T) {
		input := map[string]bool{
			"success": true,
			"failure": false,
		}

		result := TransformMap(input, func(shouldSucceed bool) func() (string, error) {
			return func() (string, error) {
				if shouldSucceed {
					return "success", nil
				}
				return "", errors.New("operation failed")
			}
		})

		// Test success case
		successResult, successErr := result["success"]()
		if successErr != nil {
			t.Errorf("success operation failed: %v", successErr)
		}
		if successResult != "success" {
			t.Errorf("success operation returned wrong value: %s", successResult)
		}

		// Test failure case
		failureResult, failureErr := result["failure"]()
		if failureErr == nil {
			t.Error("failure operation should have returned an error")
		}
		if failureResult != "" {
			t.Errorf("failure operation should return empty string, got: %s", failureResult)
		}
	})

	t.Run("empty map", func(t *testing.T) {
		input := map[string]int{}

		result := TransformMap(input, func(i int) string {
			return string(rune(i))
		})

		if len(result) != 0 {
			t.Errorf("expected empty result, got %d items", len(result))
		}
	})

	t.Run("preserve key types", func(t *testing.T) {
		input := map[int]string{
			1: "one",
			2: "two",
			3: "three",
		}

		result := TransformMap(input, func(s string) int {
			switch s {
			case "one":
				return 1
			case "two":
				return 2
			case "three":
				return 3
			default:
				return 0
			}
		})

		expected := map[int]int{
			1: 1,
			2: 2,
			3: 3,
		}

		if len(result) != len(expected) {
			t.Fatalf("expected %d results, got %d", len(expected), len(result))
		}

		for key, expectedValue := range expected {
			if actualValue, exists := result[key]; !exists {
				t.Errorf("missing key %d", key)
			} else if actualValue != expectedValue {
				t.Errorf("key %d: expected %d, got %d", key, expectedValue, actualValue)
			}
		}
	})

	t.Run("complex transformation", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		input := map[string]Person{
			"alice": {Name: "Alice", Age: 30},
			"bob":   {Name: "Bob", Age: 25},
		}

		result := TransformMap(input, func(p Person) func() string {
			return func() string {
				return p.Name + " is " + strconv.Itoa(p.Age) + " years old"
			}
		})

		aliceResult := result["alice"]()
		if aliceResult != "Alice is 30 years old" {
			t.Errorf("alice transformation failed: got %s", aliceResult)
		}
	})
}

func TestToString(t *testing.T) {
	t.Run("convert custom string types", func(t *testing.T) {
		type CustomString string
		input := []CustomString{"hello", "world", "test"}

		result := ToString(input)

		expected := []string{"hello", "world", "test"}
		if len(result) != len(expected) {
			t.Fatalf("expected %d results, got %d", len(expected), len(result))
		}

		for i, expected := range expected {
			if result[i] != expected {
				t.Errorf("index %d: expected %s, got %s", i, expected, result[i])
			}
		}
	})

	t.Run("convert empty slice", func(t *testing.T) {
		type CustomString string
		input := []CustomString{}

		result := ToString(input)

		if len(result) != 0 {
			t.Errorf("expected empty result, got %d items", len(result))
		}
	})

	t.Run("convert types.Language", func(t *testing.T) {
		// Simulate types.Language behavior
		type Language string
		input := []Language{"go", "python", "javascript"}

		result := ToString(input)

		expected := []string{"go", "python", "javascript"}
		if len(result) != len(expected) {
			t.Fatalf("expected %d results, got %d", len(expected), len(result))
		}

		for i, expected := range expected {
			if result[i] != expected {
				t.Errorf("index %d: expected %s, got %s", i, expected, result[i])
			}
		}
	})
}

func BenchmarkTransformMap(b *testing.B) {
	input := make(map[int]int)
	for i := 0; i < 1000; i++ {
		input[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TransformMap(input, func(x int) int {
			return x * 2
		})
	}
}

func BenchmarkToString(b *testing.B) {
	type CustomString string
	input := make([]CustomString, 1000)
	for i := 0; i < 1000; i++ {
		input[i] = CustomString(fmt.Sprintf("item%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToString(input)
	}
}