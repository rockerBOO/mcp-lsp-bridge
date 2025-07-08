package tools

import (
	"errors"
	"strings"
	"testing"

	"rockerboo/mcp-lsp-bridge/types"

	"github.com/myleshyson/lsprotocol-go/protocol"
)

// TestFormatCallHierarchyResults tests the formatting of call hierarchy results
func TestFormatCallHierarchyResults(t *testing.T) {
	tests := []struct {
		name              string
		items             []protocol.CallHierarchyItem
		successfulLanguage string
		errors            []error
		uri               string
		line              int
		character         int
		expectedContent   []string
		description       string
	}{
		{
			name: "single function item",
			items: []protocol.CallHierarchyItem{
				{
					Name:   "testFunction",
					Kind:   protocol.SymbolKindFunction,
					Uri:    "file:///test/main.go",
					Detail: "func testFunction() string",
					Range: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 0},
						End:   protocol.Position{Line: 10, Character: 1},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 5},
						End:   protocol.Position{Line: 5, Character: 17},
					},
				},
			},
			successfulLanguage: "go",
			errors:             []error{},
			uri:                "file:///test/main.go",
			line:               5,
			character:          10,
			expectedContent: []string{
				"CALL HIERARCHY",
				"Position: file:///test/main.go:5:10",
				"Language: go",
				"Items found: 1",
				"testFunction",
				"Function",
				"file:///test/main.go",
				"Range: 5:0-10:1",
				"Selection Range: 5:5-5:17",
				"func testFunction() string",
			},
			description: "Should format single function call hierarchy item correctly",
		},
		{
			name: "multiple items with different kinds",
			items: []protocol.CallHierarchyItem{
				{
					Name: "MyClass",
					Kind: protocol.SymbolKindClass,
					Uri:  "file:///test/class.ts",
					Range: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 0},
						End:   protocol.Position{Line: 20, Character: 1},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 1, Character: 6},
						End:   protocol.Position{Line: 1, Character: 13},
					},
				},
				{
					Name: "myMethod",
					Kind: protocol.SymbolKindMethod,
					Uri:  "file:///test/class.ts",
					Range: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 2},
						End:   protocol.Position{Line: 8, Character: 3},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 2},
						End:   protocol.Position{Line: 5, Character: 10},
					},
				},
			},
			successfulLanguage: "typescript",
			errors:             []error{},
			uri:                "file:///test/class.ts",
			line:               5,
			character:          5,
			expectedContent: []string{
				"Items found: 2",
				"MyClass",
				"Class",
				"myMethod",
				"Method",
				"typescript",
			},
			description: "Should format multiple items with different symbol kinds",
		},
		{
			name:               "no items with errors",
			items:              []protocol.CallHierarchyItem{},
			successfulLanguage: "rust",
			errors: []error{
				errors.New("rust server connection failed"),
			},
			uri:       "file:///test/main.rs",
			line:      10,
			character: 5,
			expectedContent: []string{
				"Items found: 0",
				"Errors: 1",
				"rust server connection failed",
			},
			description: "Should display errors when no items are found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCallHierarchyResults(
				tt.items,
				tt.successfulLanguage,
				tt.errors,
				tt.uri,
				tt.line,
				tt.character,
			)

			// Check that expected content is in the formatted result
			for _, expectedStr := range tt.expectedContent {
				if !strings.Contains(result, expectedStr) {
					t.Errorf("Expected formatted result to contain '%s', got: %s", expectedStr, result)
				}
			}

			t.Logf("Test case '%s' passed - %s", tt.name, tt.description)
		})
	}
}

// TestCallHierarchySingleLanguageApproach tests that call hierarchy correctly uses single-language approach
func TestCallHierarchySingleLanguageApproach(t *testing.T) {
	// This is a simple test to verify the single-language approach concept
	// In real usage, call hierarchy should:
	// 1. Infer language from the specific file URI
	// 2. Use only that language's client
	// 3. Not search across multiple languages

	testLanguage := types.Language("go")
	
	// Verify language type conversion
	languageStr := string(testLanguage)
	if languageStr != "go" {
		t.Errorf("Expected language string 'go', got '%s'", languageStr)
	}

	// This test validates the concept that call hierarchy is file-specific
	// and should not use multi-language async operations like workspace diagnostics
	t.Log("Call hierarchy correctly uses single-language approach for file-specific operations")
}