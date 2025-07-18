package tools

import (
	"strings"
	"testing"

	"rockerboo/mcp-lsp-bridge/mocks"
	"rockerboo/mcp-lsp-bridge/types"
	"rockerboo/mcp-lsp-bridge/utils"

	"github.com/myleshyson/lsprotocol-go/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test the handleFileAnalysis function directly
func TestHandleFileAnalysis(t *testing.T) {
	testCases := []struct {
		name            string
		query           string
		setupMocks      func() map[types.Language]types.LanguageClientInterface
		expectedContent string
		expectSuccess   bool
	}{
		{
			name:  "successful file analysis",
			query: "test.go",
			setupMocks: func() map[types.Language]types.LanguageClientInterface {
				mockClient := &mocks.MockLanguageClient{}
				// Use NormalizeURI to get the expected path
				expectedURI := utils.NormalizeURI("test.go")
				mockClient.On("DocumentSymbols", expectedURI).Return([]protocol.DocumentSymbol{
					{
						Name: "main",
						Kind: protocol.SymbolKindFunction,
						Range: protocol.Range{
							Start: protocol.Position{Line: 0, Character: 0},
							End:   protocol.Position{Line: 10, Character: 0},
						},
					},
				}, nil)

				return map[types.Language]types.LanguageClientInterface{
					"go": mockClient,
				}
			},
			expectedContent: "Language: go",
			expectSuccess:   true,
		},
		{
			name:  "no symbols found",
			query: "empty.go",
			setupMocks: func() map[types.Language]types.LanguageClientInterface {
				mockClient := &mocks.MockLanguageClient{}
				// Use NormalizeURI to get the expected path
				expectedURI := utils.NormalizeURI("empty.go")
				mockClient.On("DocumentSymbols", expectedURI).Return([]protocol.DocumentSymbol{}, nil)

				return map[types.Language]types.LanguageClientInterface{
					"go": mockClient,
				}
			},
			expectedContent: "could not determine language for file",
			expectSuccess:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bridge := &mocks.MockBridge{}
			clients := tc.setupMocks()
			options := make(map[string]interface{})

			var response strings.Builder

			result, err := handleFileAnalysis(bridge, clients, tc.query, options, &response)

			if tc.expectSuccess {
				require.NoError(t, err)
				assert.NotNil(t, result)
			} else {
				require.Error(t, err)
			}

			if tc.expectedContent != "" {
				assert.Contains(t, response.String(), tc.expectedContent)
			}

			// Assert mock expectations
			for _, client := range clients {
				if mockClient, ok := client.(*mocks.MockLanguageClient); ok {
					mockClient.AssertExpectations(t)
				}
			}
		})
	}
}

// Test the handlePatternAnalysis function directly
func TestHandlePatternAnalysis(t *testing.T) {
	testCases := []struct {
		name            string
		query           string
		options         map[string]interface{}
		expectedContent string
		expectSuccess   bool
	}{
		{
			name:            "error handling pattern",
			query:           "error_handling",
			options:         make(map[string]interface{}),
			expectedContent: "Pattern Type: error_handling",
			expectSuccess:   true,
		},
		{
			name:            "naming conventions pattern",
			query:           "naming_conventions",
			options:         make(map[string]interface{}),
			expectedContent: "Pattern Type: naming_conventions",
			expectSuccess:   true,
		},
		{
			name:            "architecture patterns",
			query:           "architecture_patterns",
			options:         make(map[string]interface{}),
			expectedContent: "Pattern Type: architecture_patterns",
			expectSuccess:   true,
		},
		{
			name:            "invalid pattern type",
			query:           "invalid_pattern",
			options:         make(map[string]interface{}),
			expectedContent: "unsupported pattern type",
			expectSuccess:   true,
		},
		{
			name:  "pattern type from options",
			query: "some_query",
			options: map[string]interface{}{
				"pattern_type": "error_handling",
			},
			expectedContent: "Pattern Type: error_handling",
			expectSuccess:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bridge := &mocks.MockBridge{}
			clients := make(map[types.Language]types.LanguageClientInterface)

			var response strings.Builder

			result, err := handlePatternAnalysis(bridge, clients, tc.query, tc.options, &response)

			if tc.expectSuccess {
				require.NoError(t, err)
				assert.NotNil(t, result)
			} else {
				require.Error(t, err)
			}

			if tc.expectedContent != "" {
				assert.Contains(t, response.String(), tc.expectedContent)
			}
		})
	}
}

// Test ComplexityMetrics struct and its methods
func TestComplexityMetrics(t *testing.T) {
	t.Run("complexity level categorization", func(t *testing.T) {
		testCases := []struct {
			name          string
			functions     int
			classes       int
			variables     int
			expectedLevel string
			expectedScore float64
		}{
			{
				name:          "low complexity",
				functions:     2,
				classes:       0,
				variables:     1,
				expectedLevel: "low",
				expectedScore: 5.0, // 2*2 + 0*3 + 1*1 = 5
			},
			{
				name:          "medium complexity",
				functions:     10,
				classes:       2,
				variables:     5,
				expectedLevel: "medium",
				expectedScore: 31.0, // 10*2 + 2*3 + 5*1 = 31
			},
			{
				name:          "high complexity",
				functions:     30,
				classes:       5,
				variables:     10,
				expectedLevel: "high",
				expectedScore: 85.0, // 30*2 + 5*3 + 10*1 = 85
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				symbols := []protocol.DocumentSymbol{}

				// Add functions
				for i := 0; i < tc.functions; i++ {
					symbols = append(symbols, protocol.DocumentSymbol{
						Name: "func" + string(rune(i)),
						Kind: protocol.SymbolKindFunction,
						Range: protocol.Range{
							Start: protocol.Position{Line: uint32(i), Character: 0},
							End:   protocol.Position{Line: uint32(i + 1), Character: 0},
						},
					})
				}

				// Add classes
				for i := 0; i < tc.classes; i++ {
					symbols = append(symbols, protocol.DocumentSymbol{
						Name: "class" + string(rune(i)),
						Kind: protocol.SymbolKindClass,
						Range: protocol.Range{
							Start: protocol.Position{Line: uint32(i + 100), Character: 0},
							End:   protocol.Position{Line: uint32(i + 110), Character: 0},
						},
					})
				}

				// Add variables
				for i := 0; i < tc.variables; i++ {
					symbols = append(symbols, protocol.DocumentSymbol{
						Name: "var" + string(rune(i)),
						Kind: protocol.SymbolKindVariable,
						Range: protocol.Range{
							Start: protocol.Position{Line: uint32(i + 200), Character: 0},
							End:   protocol.Position{Line: uint32(i + 200), Character: 10},
						},
					})
				}

				metrics := calculateFileComplexityFromSymbols(symbols)

				assert.Equal(t, tc.functions, metrics.FunctionCount)
				assert.Equal(t, tc.classes, metrics.ClassCount)
				assert.Equal(t, tc.variables, metrics.VariableCount)
				assert.Equal(t, tc.expectedLevel, metrics.ComplexityLevel)
				assert.InDelta(t, tc.expectedScore, metrics.ComplexityScore, 0.001)
			})
		}
	})
}

// Test edge cases and error handling
func TestAnalysisEdgeCases(t *testing.T) {
	t.Run("file analysis with no language clients", func(t *testing.T) {
		bridge := &mocks.MockBridge{}
		clients := make(map[types.Language]types.LanguageClientInterface)
		options := make(map[string]interface{})

		var response strings.Builder

		result, err := handleFileAnalysis(bridge, clients, "test.go", options, &response)

		require.NoError(t, err) // Should handle gracefully
		assert.NotNil(t, result)
		assert.Contains(t, response.String(), "could not determine language for file")
	})

	t.Run("pattern analysis with empty query", func(t *testing.T) {
		bridge := &mocks.MockBridge{}
		clients := make(map[types.Language]types.LanguageClientInterface)
		options := make(map[string]interface{})

		var response strings.Builder

		result, err := handlePatternAnalysis(bridge, clients, "", options, &response)

		require.NoError(t, err) // Should handle gracefully
		assert.NotNil(t, result)
		assert.Contains(t, response.String(), "unsupported pattern type")
	})

	t.Run("complexity calculation with nil symbols", func(t *testing.T) {
		metrics := calculateFileComplexityFromSymbols(nil)

		assert.Equal(t, 0, metrics.FunctionCount)
		assert.Equal(t, 0, metrics.ClassCount)
		assert.Equal(t, 0, metrics.VariableCount)
		assert.Equal(t, "low", metrics.ComplexityLevel)
		assert.InDelta(t, 0.0, metrics.ComplexityScore, 0.001)
		assert.Equal(t, 0, metrics.TotalLines)
	})
}
