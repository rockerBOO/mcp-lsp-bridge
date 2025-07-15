package tools

import (
	"testing"

	"rockerboo/mcp-lsp-bridge/collections"
	"rockerboo/mcp-lsp-bridge/types"

	"github.com/myleshyson/lsprotocol-go/protocol"
)

// TestExtractDiagnosticsFromWorkspaceReport tests the core diagnostic extraction functionality
func TestExtractDiagnosticsFromWorkspaceReport(t *testing.T) {
	tests := []struct {
		name             string
		report           *protocol.WorkspaceDiagnosticReport
		expectedCount    int
		expectedMessages []string
	}{
		{
			name: "empty report",
			report: &protocol.WorkspaceDiagnosticReport{
				Items: []protocol.WorkspaceDocumentDiagnosticReport{},
			},
			expectedCount:    0,
			expectedMessages: []string{},
		},
		{
			name: "report with full diagnostics",
			report: &protocol.WorkspaceDiagnosticReport{
				Items: []protocol.WorkspaceDocumentDiagnosticReport{
					{
						Value: protocol.WorkspaceFullDocumentDiagnosticReport{
							Kind: "full",
							Uri:  "file:///test.go",
							Items: []protocol.Diagnostic{
								{
									Message: "test error",
									Severity: func() *protocol.DiagnosticSeverity {
										s := protocol.DiagnosticSeverityError
										return &s
									}(),
								},
								{
									Message: "test warning",
									Severity: func() *protocol.DiagnosticSeverity {
										s := protocol.DiagnosticSeverityWarning
										return &s
									}(),
								},
							},
						},
					},
				},
			},
			expectedCount:    2,
			expectedMessages: []string{"test error", "test warning"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDiagnosticsFromWorkspaceReport(tt.report)

			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d diagnostics, got %d", tt.expectedCount, len(result))
			}

			for i, expectedMsg := range tt.expectedMessages {
				if i >= len(result) {
					t.Errorf("Expected message '%s' but not enough results", expectedMsg)
					continue
				}
				if result[i].Message != expectedMsg {
					t.Errorf("Expected message '%s', got '%s'", expectedMsg, result[i].Message)
				}
			}
		})
	}
}

// TestCollectionsToString tests the ToString function used for language conversion
func TestCollectionsToString(t *testing.T) {
	tests := []struct {
		name     string
		input    []types.Language
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []types.Language{},
			expected: []string{},
		},
		{
			name: "single language",
			input: []types.Language{
				types.Language("go"),
			},
			expected: []string{"go"},
		},
		{
			name: "multiple languages",
			input: []types.Language{
				types.Language("go"),
				types.Language("typescript"),
				types.Language("rust"),
			},
			expected: []string{"go", "typescript", "rust"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collections.ToString(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
			}

			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected %s at index %d, got %s", expected, i, result[i])
				}
			}
		})
	}
}

// TestFormatWorkspaceDiagnosticsByLanguage tests the formatting function
func TestFormatWorkspaceDiagnosticsByLanguage(t *testing.T) {
	tests := []struct {
		name            string
		languageResults []LanguageDiagnosticResult
		allDiagnostics  []protocol.Diagnostic
		errors          []DiagnosticError
		expectedContent []string
	}{
		{
			name:            "no diagnostics or errors",
			languageResults: []LanguageDiagnosticResult{},
			allDiagnostics:  []protocol.Diagnostic{},
			errors:          []DiagnosticError{},
			expectedContent: []string{
				"WORKSPACE DIAGNOSTICS",
				"Languages analyzed: 0",
				"Total diagnostics: 0",
				"No diagnostics found across all languages",
			},
		},
		{
			name: "diagnostics from multiple languages",
			languageResults: []LanguageDiagnosticResult{
				{
					Language: "go",
					Diagnostics: []protocol.Diagnostic{
						{
							Message: "go error",
							Severity: func() *protocol.DiagnosticSeverity {
								s := protocol.DiagnosticSeverityError
								return &s
							}(),
						},
					},
				},
				{
					Language: "typescript",
					Diagnostics: []protocol.Diagnostic{
						{
							Message: "ts warning",
							Severity: func() *protocol.DiagnosticSeverity {
								s := protocol.DiagnosticSeverityWarning
								return &s
							}(),
						},
					},
				},
			},
			allDiagnostics: []protocol.Diagnostic{
				{Message: "go error"},
				{Message: "ts warning"},
			},
			errors: []DiagnosticError{},
			expectedContent: []string{
				"Languages analyzed: 2",
				"Total diagnostics: 2",
				"SUMMARY BY SEVERITY",
				"RESULTS BY LANGUAGE",
				"go: 1 diagnostics",
				"typescript: 1 diagnostics",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatWorkspaceDiagnosticsByLanguage(tt.languageResults, tt.allDiagnostics, tt.errors)

			for _, expectedStr := range tt.expectedContent {
				if !contains(result, expectedStr) {
					t.Errorf("Expected output to contain '%s', got: %s", expectedStr, result)
				}
			}
		})
	}
}

// Helper function since strings.Contains might not be available in test context
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (len(substr) == 0 || findInString(s, substr) >= 0)
}

func findInString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}