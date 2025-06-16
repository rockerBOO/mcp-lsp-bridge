package mcpserver

import (
	"testing"

	"rockerboo/mcp-lsp-bridge/lsp"

	"github.com/mark3labs/mcp-go/server"
)

// TestMockBridge is a mock implementation of the BridgeInterface for testing
type TestMockBridge struct {
	mockInferLanguage           func(string) (string, error)
	mockGetClientForLanguage    func(string) (any, error)
	mockGetConfig               func() *lsp.LSPServerConfig
	mockCloseAllClients         func()
	mockFindSymbolReferences    func(string, string, int32, int32, bool) ([]any, error)
	mockFindSymbolDefinitions   func(string, string, int32, int32) ([]any, error)
	mockSearchTextInWorkspace   func(string, string) ([]any, error)
	mockGetMultiLanguageClients func([]string) (map[string]any, error)
}

func (m *TestMockBridge) GetClientForLanguageInterface(language string) (any, error) {
	if m.mockGetClientForLanguage != nil {
		return m.mockGetClientForLanguage(language)
	}
	return nil, nil
}

func (m *TestMockBridge) InferLanguage(filePath string) (string, error) {
	if m.mockInferLanguage != nil {
		return m.mockInferLanguage(filePath)
	}
	return "go", nil
}

func (m *TestMockBridge) CloseAllClients() {
	if m.mockCloseAllClients != nil {
		m.mockCloseAllClients()
	}
}

func (m *TestMockBridge) GetConfig() *lsp.LSPServerConfig {
	if m.mockGetConfig != nil {
		return m.mockGetConfig()
	}
	return &lsp.LSPServerConfig{
		LanguageServers: map[string]lsp.LanguageServerConfig{
			"go": {
				Command: "gopls",
				Filetypes: []string{".go"},
			},
			"python": {
				Command: "pyright",
				Filetypes: []string{".py"},
			},
		},
		ExtensionLanguageMap: map[string]string{
			".go": "go",
				".py": "python",
		},
	}
}

func (m *TestMockBridge) DetectProjectLanguages(projectPath string) ([]string, error) {
	return []string{"go"}, nil
}

func (m *TestMockBridge) DetectPrimaryProjectLanguage(projectPath string) (string, error) {
	return "go", nil
}

func (m *TestMockBridge) FindSymbolReferences(language, uri string, line, character int32, includeDeclaration bool) ([]any, error) {
	if m.mockFindSymbolReferences != nil {
		return m.mockFindSymbolReferences(language, uri, line, character, includeDeclaration)
	}
	return []any{}, nil
}

func (m *TestMockBridge) FindSymbolDefinitions(language, uri string, line, character int32) ([]any, error) {
	if m.mockFindSymbolDefinitions != nil {
		return m.mockFindSymbolDefinitions(language, uri, line, character)
	}
	return []any{}, nil
}

func (m *TestMockBridge) SearchTextInWorkspace(language, query string) ([]any, error) {
	if m.mockSearchTextInWorkspace != nil {
		return m.mockSearchTextInWorkspace(language, query)
	}
	return []any{}, nil
}

func (m *TestMockBridge) GetMultiLanguageClients(languages []string) (map[string]any, error) {
	if m.mockGetMultiLanguageClients != nil {
		return m.mockGetMultiLanguageClients(languages)
	}
	result := make(map[string]any)
	for _, lang := range languages {
		if lang == "go" {
			result[lang] = &lsp.LanguageClient{}
		}
	}
	return result, nil
}

func TestMCPServerToolsSetup(t *testing.T) {
	testCases := []struct {
		name             string
		toolRegistration func(*server.MCPServer, *TestMockBridge)
	}{
		{
			name: "Analyze Code Tool",
			toolRegistration: func(mcpServer *server.MCPServer, mockBridge *TestMockBridge) {
				registerAnalyzeCodeTool(mcpServer, mockBridge)
			},
		},
		{
			name: "Infer Language Tool",
			toolRegistration: func(mcpServer *server.MCPServer, mockBridge *TestMockBridge) {
				registerInferLanguageTool(mcpServer, mockBridge)
			},
		},
		{
			name: "LSP Connect Tool",
			toolRegistration: func(mcpServer *server.MCPServer, mockBridge *TestMockBridge) {
				registerLSPConnectTool(mcpServer, mockBridge)
			},
		},
		{
			name: "LSP Disconnect Tool",
			toolRegistration: func(mcpServer *server.MCPServer, mockBridge *TestMockBridge) {
				registerLSPDisconnectTool(mcpServer, mockBridge)
			},
		},
		{
			name: "Detect Project Languages Tool",
			toolRegistration: func(mcpServer *server.MCPServer, mockBridge *TestMockBridge) {
				registerProjectLanguageDetectionTool(mcpServer, mockBridge)
			},
		},
		{
			name: "Project Analysis Tool",
			toolRegistration: func(mcpServer *server.MCPServer, mockBridge *TestMockBridge) {
				registerProjectAnalysisTool(mcpServer, mockBridge)
			},
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare mock bridge
			mockBridge := &TestMockBridge{
				mockInferLanguage: func(path string) (string, error) {
					if path == "/test/example.go" || path == "/test/file.go" {
						return "go", nil
					}
					return "unknown", nil
				},
				mockGetClientForLanguage: func(language string) (any, error) {
					if language == "go" {
						return &lsp.LanguageClient{}, nil
					}
					return nil, nil
				},
				mockGetConfig: func() *lsp.LSPServerConfig {
					return &lsp.LSPServerConfig{
						LanguageServers: map[string]lsp.LanguageServerConfig{
							"go": {Command: "gopls"},
						},
					}
				},
				mockCloseAllClients: func() {},
			}

			// Create MCP server
			mcpServer := server.NewMCPServer(
				"lsp-bridge-mcp",
				"1.0.0",
				server.WithToolCapabilities(false),
			)

			// Register tools
			tc.toolRegistration(mcpServer, mockBridge)

			// No verification needed - test will pass if registration doesn't panic
		})
	}
}

// Test enhanced bridge functionality
func TestEnhancedBridgeMethods(t *testing.T) {
	testCases := []struct {
		name     string
		testFunc func(t *testing.T, bridge *TestMockBridge)
	}{
		{
			name: "FindSymbolReferences",
			testFunc: func(t *testing.T, bridge *TestMockBridge) {
				bridge.mockFindSymbolReferences = func(language, uri string, line, character int32, includeDeclaration bool) ([]any, error) {
					if language == "go" && uri == "file:///test.go" {
						return []any{"reference1", "reference2"}, nil
					}
					return []any{}, nil
				}

				refs, err := bridge.FindSymbolReferences("go", "file:///test.go", 10, 5, true)
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if len(refs) != 2 {
					t.Errorf("Expected 2 references, got %d", len(refs))
				}
			},
		},
		{
			name: "FindSymbolDefinitions",
			testFunc: func(t *testing.T, bridge *TestMockBridge) {
				bridge.mockFindSymbolDefinitions = func(language, uri string, line, character int32) ([]any, error) {
					if language == "go" && uri == "file:///test.go" {
						return []any{"definition1"}, nil
					}
					return []any{}, nil
				}

				defs, err := bridge.FindSymbolDefinitions("go", "file:///test.go", 10, 5)
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if len(defs) != 1 {
					t.Errorf("Expected 1 definition, got %d", len(defs))
				}
			},
		},
		{
			name: "SearchTextInWorkspace",
			testFunc: func(t *testing.T, bridge *TestMockBridge) {
				bridge.mockSearchTextInWorkspace = func(language, query string) ([]any, error) {
					if language == "go" && query == "main" {
						return []any{"main function", "main package"}, nil
					}
					return []any{}, nil
				}

				results, err := bridge.SearchTextInWorkspace("go", "main")
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if len(results) != 2 {
					t.Errorf("Expected 2 search results, got %d", len(results))
				}
			},
		},
		{
			name: "GetMultiLanguageClients",
			testFunc: func(t *testing.T, bridge *TestMockBridge) {
				bridge.mockGetMultiLanguageClients = func(languages []string) (map[string]any, error) {
					result := make(map[string]any)
					for _, lang := range languages {
						if lang == "go" || lang == "python" {
							result[lang] = &lsp.LanguageClient{}
						}
					}
					return result, nil
				}

				clients, err := bridge.GetMultiLanguageClients([]string{"go", "python", "rust"})
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if len(clients) != 2 {
					t.Errorf("Expected 2 clients, got %d", len(clients))
				}
				if _, exists := clients["go"]; !exists {
					t.Error("Expected go client to exist")
				}
				if _, exists := clients["python"]; !exists {
					t.Error("Expected python client to exist")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockBridge := &TestMockBridge{}
			tc.testFunc(t, mockBridge)
		})
	}
}

// Test project analysis tool functionality
func TestProjectAnalysisToolEnhancements(t *testing.T) {
	mockBridge := &TestMockBridge{
		mockGetMultiLanguageClients: func(languages []string) (map[string]any, error) {
			result := make(map[string]any)
			for _, lang := range languages {
				if lang == "go" {
					// Create a mock language client with necessary methods
					client := &MockLanguageClient{}
					result[lang] = client
				}
			}
			return result, nil
		},
		mockFindSymbolReferences: func(language, uri string, line, character int32, includeDeclaration bool) ([]any, error) {
			return []any{"mock reference"}, nil
		},
		mockFindSymbolDefinitions: func(language, uri string, line, character int32) ([]any, error) {
			return []any{"mock definition"}, nil
		},
		mockSearchTextInWorkspace: func(language, query string) ([]any, error) {
			return []any{"mock search result"}, nil
		},
	}

	mcpServer := server.NewMCPServer(
		"test-server",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Register the enhanced project analysis tool
	registerProjectAnalysisTool(mcpServer, mockBridge)

	// Verify that the tool was registered without errors
	// Additional integration testing would require MCP server test utilities
}

// MockLanguageClient for testing
type MockLanguageClient struct{}

func (m *MockLanguageClient) WorkspaceSymbols(query string) ([]any, error) {
	// Return mock symbol information with location data
	mockSymbol := struct {
		Name     string
		Location struct {
			Uri   string
			Range struct {
				Start struct {
					Line      int32
					Character int32
				}
			}
		}
	}{
		Name: "mockSymbol",
		Location: struct {
			Uri   string
			Range struct {
				Start struct {
					Line      int32
					Character int32
				}
			}
		}{
			Uri: "file:///test.go",
			Range: struct {
				Start struct {
					Line      int32
					Character int32
				}
			}{
				Start: struct {
					Line      int32
					Character int32
				}{
					Line:      10,
					Character: 5,
				},
			},
		},
	}
	return []any{mockSymbol}, nil
}

// Benchmark server setup performance
func BenchmarkMCPServerToolRegistration(b *testing.B) {
	mockBridge := &TestMockBridge{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SetupMCPServer(mockBridge)
	}
}
