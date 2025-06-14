package mcpserver

import (
	"testing"

	"rockerboo/mcp-lsp-bridge/lsp"

	"github.com/mark3labs/mcp-go/server"
)

// TestMockBridge is a mock implementation of the BridgeInterface for testing
type TestMockBridge struct {
	mockInferLanguage        func(string) (string, error)
	mockGetClientForLanguage func(string) (any, error)
	mockGetConfig            func() *lsp.LSPServerConfig
	mockCloseAllClients     func()
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
			"go": {Command: "gopls"},
		},
	}
}

func TestMCPServerToolsSetup(t *testing.T) {
	testCases := []struct {
		name            string
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

// Benchmark server setup performance
func BenchmarkMCPServerToolRegistration(b *testing.B) {
	mockBridge := &TestMockBridge{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SetupMCPServer(mockBridge)
	}
}