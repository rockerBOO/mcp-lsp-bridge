package mcp_lsp_bridge_test

import(
	"rockerboo/mcp-lsp-bridge/mcp_lsp_bridge"
	"testing"
)

func TestPrettyPrint(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		v    any
		want string
	}{
		{"nil", nil, "null"},
		{"int", 123, "123"},
		{"string", "hello", "\"hello\""},
		{"bool", true, "true"},
		{"slice", []int{1, 2, 3}, "[\n   1,\n   2,\n   3\n]"},
		{"map", map[string]int{"a": 1, "b": 2}, "{\n   \"a\": 1,\n   \"b\": 2\n}"},
		{"struct", struct{ A int; B string }{1, "hello"}, "{\n   \"A\": 1,\n   \"B\": \"hello\"\n}"},
		{"slice of structs", []struct{ A int; B string }{{1, "hello"}, {2, "world"}}, "[\n   {\n      \"A\": 1,\n      \"B\": \"hello\"\n   },\n   {\n      \"A\": 2,\n      \"B\": \"world\"\n   }\n]"},
		{"map of slices", map[string][]int{"a": {1, 2, 3}, "b": {4, 5, 6}}, "{\n   \"a\": [\n      1,\n      2,\n      3\n   ],\n   \"b\": [\n      4,\n      5,\n      6\n   ]\n}"},
		{"slice of maps", []map[string]int{{"a": 1, "b": 2}, {"c": 3, "d": 4}}, "[\n   {\n      \"a\": 1,\n      \"b\": 2\n   },\n   {\n      \"c\": 3,\n      \"d\": 4\n   }\n]"},
		{"map of structs", map[string]struct{ A int; B string }{"a": {1, "hello"}, "b": {2, "world"}}, "{\n   \"a\": {\n      \"A\": 1,\n      \"B\": \"hello\"\n   },\n   \"b\": {\n      \"A\": 2,\n      \"B\": \"world\"\n   }\n}"},
		{"slice of maps of structs", []map[string]struct{ A int; B string }{{"a": {1, "hello"}, "b": {2, "world"}}, {"c": {3, "foo"}, "d": {4, "bar"}}}, "[\n   {\n      \"a\": {\n         \"A\": 1,\n         \"B\": \"hello\"\n      },\n      \"b\": {\n         \"A\": 2,\n         \"B\": \"world\"\n      }\n   },\n   {\n      \"c\": {\n         \"A\": 3,\n         \"B\": \"foo\"\n      },\n      \"d\": {\n         \"A\": 4,\n         \"B\": \"bar\"\n      }\n   }\n]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mcp_lsp_bridge.PrettyPrint(tt.v)
			if got != tt.want {
				t.Errorf("PrettyPrint() = %v, want %v", got, tt.want)
			}
		})
	}
}
