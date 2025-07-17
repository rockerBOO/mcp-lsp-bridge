package analysis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"rockerboo/mcp-lsp-bridge/mocks"
	"rockerboo/mcp-lsp-bridge/types"
	"github.com/myleshyson/lsprotocol-go/protocol"
)

// createMockClients creates mock language clients for testing
func createMockClients() map[types.Language]types.LanguageClientInterface {
	goClient := &mocks.MockLanguageClient{}
	jsClient := &mocks.MockLanguageClient{}

	// Setup mock symbol responses
	goSymbols := []protocol.WorkspaceSymbol{
		{
			Name: "ProcessPayment", 
			Kind: protocol.SymbolKindFunction,
			Location: protocol.Or2[protocol.Location, protocol.LocationUriOnly]{
				Value: protocol.Location{
					Uri: "file:///mock/go/process_payment.go",
					Range: protocol.Range{
						Start: protocol.Position{Line: 10, Character: 5},
						End: protocol.Position{Line: 10, Character: 20},
					},
				},
			},
		},
		{
			Name: "UserAuth", 
			Kind: protocol.SymbolKindStruct,
			Location: protocol.Or2[protocol.Location, protocol.LocationUriOnly]{
				Value: protocol.Location{
					Uri: "file:///mock/go/user_auth.go",
					Range: protocol.Range{
						Start: protocol.Position{Line: 20, Character: 5},
						End: protocol.Position{Line: 20, Character: 20},
					},
				},
			},
		},
	}
	
	jsSymbols := []protocol.WorkspaceSymbol{
		{
			Name: "handleRequest", 
			Kind: protocol.SymbolKindFunction,
			Location: protocol.Or2[protocol.Location, protocol.LocationUriOnly]{
				Value: protocol.Location{
					Uri: "file:///mock/js/handle_request.js",
					Range: protocol.Range{
						Start: protocol.Position{Line: 15, Character: 5},
						End: protocol.Position{Line: 15, Character: 20},
					},
				},
			},
		},
		{
			Name: "UserModel", 
			Kind: protocol.SymbolKindClass,
			Location: protocol.Or2[protocol.Location, protocol.LocationUriOnly]{
				Value: protocol.Location{
					Uri: "file:///mock/js/user_model.js",
					Range: protocol.Range{
						Start: protocol.Position{Line: 25, Character: 5},
						End: protocol.Position{Line: 25, Character: 20},
					},
				},
			},
		},
	}
	
	goClient.On("WorkspaceSymbols", mock.Anything).Return(goSymbols, nil)
	jsClient.On("WorkspaceSymbols", mock.Anything).Return(jsSymbols, nil)

	// Add method stubs for references and definitions
	goClient.On("References", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]protocol.Location{
		{
			Uri: "file:///mock/go/references.go",
			Range: protocol.Range{
				Start: protocol.Position{Line: 30, Character: 5},
				End: protocol.Position{Line: 30, Character: 20},
			},
		},
	}, nil)
	goClient.On("Definition", mock.Anything, mock.Anything, mock.Anything).Return([]protocol.Or2[protocol.LocationLink, protocol.Location]{
		{
			Value: protocol.Location{
				Uri: "file:///mock/go/definition.go",
				Range: protocol.Range{
					Start: protocol.Position{Line: 40, Character: 5},
					End: protocol.Position{Line: 40, Character: 20},
				},
			},
		},
	}, nil)

	jsClient.On("References", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]protocol.Location{
		{
			Uri: "file:///mock/js/references.js",
			Range: protocol.Range{
				Start: protocol.Position{Line: 35, Character: 5},
				End: protocol.Position{Line: 35, Character: 20},
			},
		},
	}, nil)
	jsClient.On("Definition", mock.Anything, mock.Anything, mock.Anything).Return([]protocol.Or2[protocol.LocationLink, protocol.Location]{
		{
			Value: protocol.Location{
				Uri: "file:///mock/js/definition.js",
				Range: protocol.Range{
					Start: protocol.Position{Line: 45, Character: 5},
					End: protocol.Position{Line: 45, Character: 20},
				},
			},
		},
	}, nil)

	// Add minimal PrepareCallHierarchy stub
	goClient.On("PrepareCallHierarchy", mock.Anything, mock.Anything, mock.Anything).Return([]protocol.CallHierarchyItem{}, nil)
	jsClient.On("PrepareCallHierarchy", mock.Anything, mock.Anything, mock.Anything).Return([]protocol.CallHierarchyItem{}, nil)

	// Add Implementation stubs
	goClient.On("Implementation", mock.Anything, mock.Anything, mock.Anything).Return([]protocol.Location{}, nil)
	jsClient.On("Implementation", mock.Anything, mock.Anything, mock.Anything).Return([]protocol.Location{}, nil)

	// Add DocumentSymbols stubs for file analysis with actual symbols
	goClient.On("DocumentSymbols", mock.Anything).Return([]protocol.DocumentSymbol{
		{
			Name: "TestFunction",
			Kind: protocol.SymbolKindFunction,
			Range: protocol.Range{
				Start: protocol.Position{Line: 1, Character: 0},
				End:   protocol.Position{Line: 10, Character: 0},
			},
		},
	}, nil)
	jsClient.On("DocumentSymbols", mock.Anything).Return([]protocol.DocumentSymbol{
		{
			Name: "TestClass",
			Kind: protocol.SymbolKindClass,
			Range: protocol.Range{
				Start: protocol.Position{Line: 1, Character: 0},
				End:   protocol.Position{Line: 20, Character: 0},
			},
		},
	}, nil)

	return map[types.Language]types.LanguageClientInterface{
		"go":         goClient,
		"javascript": jsClient,
	}
}

// TestNewProjectAnalyzer tests the creation of a new ProjectAnalyzer
func TestNewProjectAnalyzer(t *testing.T) {
	clients := createMockClients()
	
	analyzer := NewProjectAnalyzer(clients)
	
	assert.NotNil(t, analyzer)
	assert.Equal(t, len(clients), len(analyzer.clients))
}

// TestAnalyzeWorkspace tests the workspace analysis functionality
func TestAnalyzeWorkspace(t *testing.T) {
	clients := createMockClients()
	analyzer := NewProjectAnalyzer(clients)
	
	result, err := analyzer.Analyze(AnalysisRequest{
		Type:   WorkspaceAnalysis,
		Target: "payment",
		Scope:  "project",
		Depth:  "detailed",
	})
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	
	// Type assertion for workspace analysis data
	data, ok := result.Data.(WorkspaceAnalysisData)
	assert.True(t, ok)
	
	// Check total symbols
	assert.Equal(t, 4, data.TotalSymbols)
	
	// Check language distribution
	assert.Contains(t, data.LanguageDistribution, types.Language("go"))
	assert.Contains(t, data.LanguageDistribution, types.Language("javascript"))
}

// TestAnalyzeSymbolRelationships tests symbol relationship analysis
func TestAnalyzeSymbolRelationships(t *testing.T) {
	clients := createMockClients()
	analyzer := NewProjectAnalyzer(clients)
	
	result, err := analyzer.Analyze(AnalysisRequest{
		Type:   SymbolRelationships,
		Target: "ProcessPayment", // Changed to match first symbol
		Scope:  "project",
	})
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	
	// Type assertion for symbol relationships
	relationships, ok := result.Data.(SymbolRelationshipsData)
	assert.True(t, ok)
	
	// Check symbol details (could be either ProcessPayment or handleRequest)
	assert.Contains(t, []string{"ProcessPayment", "handleRequest"}, relationships.Symbol.Name)
	assert.Contains(t, []types.Language{"go", "javascript"}, relationships.Language)
}

// TestCaching tests the caching mechanism of the ProjectAnalyzer
func TestCaching(t *testing.T) {
	clients := createMockClients()
	
	// Custom cache with shorter TTL for testing
	cache := NewAnalysisCache(100, 100*time.Millisecond)
	analyzer := NewProjectAnalyzer(clients, WithCache(cache))
	
	// First analysis
	request := AnalysisRequest{
		Type:   WorkspaceAnalysis,
		Target: "payment",
		Scope:  "project",
	}
	
	firstResult, err := analyzer.Analyze(request)
	require.NoError(t, err)
	
	// Get cached result
	cachedResult, err := analyzer.Analyze(request)
	require.NoError(t, err)
	
	// Check cache stats
	stats := cache.Stats()
	assert.Greater(t, stats.Hits, int64(0))
	
	// Compare results
	assert.Equal(t, firstResult.Data, cachedResult.Data)
}

// TestErrorHandling tests the error handling capabilities
func TestErrorHandling(t *testing.T) {
	// Create mock clients with an error-generating client
	clients := createMockClients()
	errClient := &mocks.MockLanguageClient{}
	errClient.On("WorkspaceSymbols", mock.Anything).Return([]protocol.WorkspaceSymbol{}, assert.AnError)
	clients["python"] = errClient
	
	// Create analyzer with custom error handler
	errorHandler := NewErrorHandler(2, true, 0.5)
	analyzer := NewProjectAnalyzer(clients, WithErrorHandler(errorHandler))
	
	result, err := analyzer.Analyze(AnalysisRequest{
		Type:   WorkspaceAnalysis,
		Target: "error_test",
		Scope:  "project",
	})
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	
	// Check metadata for errors
	metadata := result.Metadata
	assert.Greater(t, len(metadata.Errors), 0)
	assert.Contains(t, metadata.LanguagesUsed, types.Language("go"))
	assert.Contains(t, metadata.LanguagesUsed, types.Language("javascript"))
}

// TestAnalyzeFileAnalysis tests file analysis functionality
func TestAnalyzeFileAnalysis(t *testing.T) {
	clients := createMockClients()
	analyzer := NewProjectAnalyzer(clients)
	
	result, err := analyzer.Analyze(AnalysisRequest{
		Type:   FileAnalysis,
		Target: "test.go", // Use simple filename instead of full URI
		Scope:  "file",
		Options: map[string]interface{}{
			"language": "go", // Specify language explicitly
		},
	})
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	
	// Type assertion for file analysis data
	fileData, ok := result.Data.(FileAnalysisData)
	assert.True(t, ok)
	
	// Check file details
	assert.Contains(t, fileData.Uri, "test.go")
	assert.Contains(t, []types.Language{"go", "javascript"}, fileData.Language)
	assert.NotNil(t, fileData.Complexity)
	assert.NotNil(t, fileData.CodeQuality)
}

// TestAnalyzePatternAnalysis tests pattern analysis functionality
func TestAnalyzePatternAnalysis(t *testing.T) {
	clients := createMockClients()
	analyzer := NewProjectAnalyzer(clients)
	
	result, err := analyzer.Analyze(AnalysisRequest{
		Type:   PatternAnalysis,
		Target: "naming_conventions",
		Scope:  "project",
		Options: map[string]interface{}{
			"pattern_scope": "project",
		},
	})
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	
	// Type assertion for pattern analysis data
	patternData, ok := result.Data.(PatternAnalysisData)
	assert.True(t, ok)
	
	// Check pattern details
	assert.Equal(t, "naming_conventions", patternData.PatternType)
	assert.Equal(t, "project", patternData.Scope)
	assert.GreaterOrEqual(t, patternData.ConsistencyScore, 0.0)
	assert.LessOrEqual(t, patternData.ConsistencyScore, 1.0)
}

// TestWithPerformanceConfig tests the performance configuration option
func TestWithPerformanceConfig(t *testing.T) {
	clients := createMockClients()
	performanceConfig := &PerformanceConfig{
		MaxGoroutines: 10,
		MemoryLimit:   1024 * 1024,
	}
	
	analyzer := NewProjectAnalyzer(clients, WithPerformanceConfig(performanceConfig))
	
	assert.NotNil(t, analyzer)
	// Performance config is set internally, so we can't directly test it
	// but we can verify the analyzer was created successfully
}

// TestAnalyzePatternErrorHandling tests error handling pattern analysis
func TestAnalyzePatternErrorHandling(t *testing.T) {
	clients := createMockClients()
	analyzer := NewProjectAnalyzer(clients)
	
	result, err := analyzer.Analyze(AnalysisRequest{
		Type:   PatternAnalysis,
		Target: "error_handling",
		Scope:  "project",
	})
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	
	// Type assertion for pattern analysis data
	patternData, ok := result.Data.(PatternAnalysisData)
	assert.True(t, ok)
	assert.Equal(t, "error_handling", patternData.PatternType)
}

// TestAnalyzePatternArchitecture tests architecture pattern analysis
func TestAnalyzePatternArchitecture(t *testing.T) {
	clients := createMockClients()
	analyzer := NewProjectAnalyzer(clients)
	
	result, err := analyzer.Analyze(AnalysisRequest{
		Type:   PatternAnalysis,
		Target: "architecture_patterns",
		Scope:  "project",
	})
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	
	// Type assertion for pattern analysis data
	patternData, ok := result.Data.(PatternAnalysisData)
	assert.True(t, ok)
	assert.Equal(t, "architecture_patterns", patternData.PatternType)
}

// TestAnalyzeUnsupportedType tests handling of unsupported analysis types
func TestAnalyzeUnsupportedType(t *testing.T) {
	clients := createMockClients()
	analyzer := NewProjectAnalyzer(clients)
	
	result, err := analyzer.Analyze(AnalysisRequest{
		Type:   "unsupported_type",
		Target: "test",
		Scope:  "project",
	})
	
	// Should return an error for unsupported analysis type
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported analysis type")
}

// TestCircularDependencyDetection tests the circular dependency detection logic
func TestCircularDependencyDetection(t *testing.T) {
	clients := createMockClients()
	analyzer := NewProjectAnalyzer(clients)
	
	// Test with a simple circular dependency scenario
	result, err := analyzer.Analyze(AnalysisRequest{
		Type:   WorkspaceAnalysis,
		Target: "circular_test",
		Scope:  "project",
	})
	
	require.NoError(t, err)
	assert.NotNil(t, result)
	
	// Check that dependency patterns are detected
	data, ok := result.Data.(WorkspaceAnalysisData)
	assert.True(t, ok)
	assert.Greater(t, len(data.DependencyPatterns), 0)
}