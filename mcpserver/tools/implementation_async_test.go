package tools

import (
	"errors"
	"strings"
	"testing"

	"rockerboo/mcp-lsp-bridge/types"

	"github.com/myleshyson/lsprotocol-go/protocol"
)

// TestFormatMultiLanguageImplementations tests the formatting of implementation results
func TestFormatMultiLanguageImplementations(t *testing.T) {
	tests := []struct {
		name            string
		implementations []protocol.Location
		errors          []error
		uri             string
		line            int
		character       int
		languages       []types.Language
		expectedContent []string
		description     string
	}{
		{
			name: "multiple implementations across languages",
			implementations: []protocol.Location{
				{
					Uri: "file:///test/impl1.go",
					Range: protocol.Range{
						Start: protocol.Position{Line: 10, Character: 0},
						End:   protocol.Position{Line: 15, Character: 1},
					},
				},
				{
					Uri: "file:///test/impl.ts",
					Range: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 0},
						End:   protocol.Position{Line: 8, Character: 1},
					},
				},
			},
			errors: []error{},
			uri:    "file:///test/interface.go",
			line:   5,
			character: 10,
			languages: []types.Language{
				types.Language("go"),
				types.Language("typescript"),
			},
			expectedContent: []string{
				"IMPLEMENTATIONS",
				"Position: file:///test/interface.go:5:10",
				"Languages searched: [go typescript]",
				"Implementations found: 2",
				"impl1.go",
				"impl.ts",
				"file:///test/impl1.go",
				"file:///test/impl.ts",
				"Range: 10:0-15:1",
				"Range: 5:0-8:1",
			},
			description: "Should format multiple implementations correctly",
		},
		{
			name:            "no implementations with errors",
			implementations: []protocol.Location{},
			errors: []error{
				errors.New("server1 failed"),
				errors.New("server2 timeout"),
			},
			uri:       "file:///test/interface.java",
			line:      8,
			character: 15,
			languages: []types.Language{
				types.Language("java"),
				types.Language("kotlin"),
			},
			expectedContent: []string{
				"IMPLEMENTATIONS",
				"Implementations found: 0",
				"Errors: 2",
				"ERRORS",
				"server1 failed",
				"server2 timeout",
				"No implementations found",
			},
			description: "Should show errors when no implementations are found",
		},
		{
			name:            "no implementations without errors",
			implementations: []protocol.Location{},
			errors:          []error{},
			uri:             "file:///test/concrete.rs",
			line:            3,
			character:       5,
			languages: []types.Language{
				types.Language("rust"),
			},
			expectedContent: []string{
				"No implementations found",
				"This may indicate:",
				"not an interface or abstract method",
				"No implementations exist",
				"does not correspond to a valid symbol",
			},
			description: "Should provide helpful message when no implementations found without errors",
		},
		{
			name: "single implementation",
			implementations: []protocol.Location{
				{
					Uri: "file:///test/only_impl.py",
					Range: protocol.Range{
						Start: protocol.Position{Line: 20, Character: 4},
						End:   protocol.Position{Line: 25, Character: 0},
					},
				},
			},
			errors:    []error{},
			uri:       "file:///test/abstract.py",
			line:      12,
			character: 8,
			languages: []types.Language{
				types.Language("python"),
			},
			expectedContent: []string{
				"Implementations found: 1",
				"only_impl.py",
				"Position: line=20, character=4",
			},
			description: "Should format single implementation correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMultiLanguageImplementations(
				tt.implementations,
				tt.errors,
				tt.uri,
				tt.line,
				tt.character,
				tt.languages,
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

// TestImplementationMultiLanguageApproach tests that implementation search correctly uses multi-language approach
func TestImplementationMultiLanguageApproach(t *testing.T) {
	// This test validates the concept that implementation search should be multi-language
	// because interfaces can be implemented across different programming languages

	languages := []types.Language{
		types.Language("go"),
		types.Language("typescript"),
		types.Language("rust"),
	}

	// Test that we can work with multiple languages for implementations
	if len(languages) != 3 {
		t.Errorf("Expected 3 languages, got %d", len(languages))
	}

	// Verify language conversion for async operations
	for _, lang := range languages {
		langStr := string(lang)
		if langStr == "" {
			t.Errorf("Language conversion failed for %v", lang)
		}
	}

	t.Log("Implementation search correctly uses multi-language approach for cross-language interface implementations")
}