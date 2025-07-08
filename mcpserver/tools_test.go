package mcpserver

import (
	"testing"

	"rockerboo/mcp-lsp-bridge/mocks"

	"github.com/mark3labs/mcp-go/server"
)

// Tests that all tools can be registered without errors
func TestRegisterAllTools(t *testing.T) {
	// mockBridge := &mocks.MockBridge{}
	//
	// // Set up expectations
	// mockBridge.On("InferLanguage", "/test/example.go").Return("go", nil)
	// mockBridge.On("InferLanguage", "/test/file.go").Return("go", nil)
	// mockBridge.On("InferLanguage", mock.AnythingOfType("string")).Return("unknown", nil)
	//
	// mockBridge.On("GetClientForLanguageInterface", "go").Return(&lsp.LanguageClient{}, nil)
	// mockBridge.On("GetClientForLanguageInterface", mock.AnythingOfType("string")).Return(nil, nil)
	//
	// mockBridge.On("GetConfig").Return(&lsp.LSPServerConfig{
	// 	LanguageServers: map[string]lsp.LanguageServerConfig{
	// 		"go": {Command: "gopls"},
	// 	},
	// })
	//
	// mockBridge.On("CloseAllClients").Return()
	//
	// // Create MCP server
	// mcpServer := server.NewMCPServer(
	// 	"lsp-bridge-mcp",
	// 	"1.0.0",
	// 	server.WithToolCapabilities(false),
	// )
	//
	// // Test that RegisterAllTools works without panicking
	// RegisterAllTools(mcpServer, mockBridge)
	//
	// // Verify expectations were met
	// mockBridge.AssertExpectations(t)

}

// TestIndividualToolRegistration tests individual tool registration functions
func TestIndividualToolRegistration(t *testing.T) {
	// testCases := []struct {
	// 	name             string
	// 	toolRegistration func(*server.MCPServer, interfaces.BridgeInterface)
	// }{
	// 	{
	// 		name: "Analyze Code Tool",
	// 		toolRegistration: func(mcpServer *server.MCPServer, bridge interfaces.BridgeInterface) {
	// 			tools.RegisterAnalyzeCodeTool(mcpServer, bridge)
	// 		},
	// 	},
	// 	{
	// 		name: "Project Analysis Tool",
	// 		toolRegistration: func(mcpServer *server.MCPServer, bridge interfaces.BridgeInterface) {
	// 			tools.RegisterProjectAnalysisTool(mcpServer, bridge)
	// 		},
	// 	},
	// 	{
	// 		name: "Infer Language Tool",
	// 		toolRegistration: func(mcpServer *server.MCPServer, bridge interfaces.BridgeInterface) {
	// 			tools.RegisterInferLanguageTool(mcpServer, bridge)
	// 		},
	// 	},
	// 	{
	// 		name: "LSP Connect Tool",
	// 		toolRegistration: func(mcpServer *server.MCPServer, bridge interfaces.BridgeInterface) {
	// 			tools.RegisterLSPConnectTool(mcpServer, bridge)
	// 		},
	// 	},
	// 	{
	// 		name: "LSP Disconnect Tool",
	// 		toolRegistration: func(mcpServer *server.MCPServer, bridge interfaces.BridgeInterface) {
	// 			tools.RegisterLSPDisconnectTool(mcpServer, bridge)
	// 		},
	// 	},
	// 	{
	// 		name: "Detect Project Languages Tool",
	// 		toolRegistration: func(mcpServer *server.MCPServer, bridge interfaces.BridgeInterface) {
	// 			tools.RegisterProjectLanguageDetectionTool(mcpServer, bridge)
	// 		},
	// 	},
	// 	{
	// 		name: "Hover Tool",
	// 		toolRegistration: func(mcpServer *server.MCPServer, bridge interfaces.BridgeInterface) {
	// 			tools.RegisterHoverTool(mcpServer, bridge)
	// 		},
	// 	},
	// 	{
	// 		name: "Diagnostics Tool",
	// 		toolRegistration: func(mcpServer *server.MCPServer, bridge interfaces.BridgeInterface) {
	// 			tools.RegisterDiagnosticsTool(mcpServer, bridge)
	// 		},
	// 	},
	// 	{
	// 		name: "Signature Help Tool",
	// 		toolRegistration: func(mcpServer *server.MCPServer, bridge interfaces.BridgeInterface) {
	// 			tools.RegisterSignatureHelpTool(mcpServer, bridge)
	// 		},
	// 	},
	// 	{
	// 		name: "Code Actions Tool",
	// 		toolRegistration: func(mcpServer *server.MCPServer, bridge interfaces.BridgeInterface) {
	// 			tools.RegisterCodeActionsTool(mcpServer, bridge)
	// 		},
	// 	},
	// 	{
	// 		name: "Format Document Tool",
	// 		toolRegistration: func(mcpServer *server.MCPServer, bridge interfaces.BridgeInterface) {
	// 			tools.RegisterFormatDocumentTool(mcpServer, bridge)
	// 		},
	// 	},
	// 	{
	// 		name: "Rename Tool",
	// 		toolRegistration: func(mcpServer *server.MCPServer, bridge interfaces.BridgeInterface) {
	// 			tools.RegisterRenameTool(mcpServer, bridge)
	// 		},
	// 	},
	// 	{
	// 		name: "Implementation Tool",
	// 		toolRegistration: func(mcpServer *server.MCPServer, bridge interfaces.BridgeInterface) {
	// 			tools.RegisterImplementationTool(mcpServer, bridge)
	// 		},
	// 	},
	// 	{
	// 		name: "Call Hierarchy Tool",
	// 		toolRegistration: func(mcpServer *server.MCPServer, bridge interfaces.BridgeInterface) {
	// 			tools.RegisterCallHierarchyTool(mcpServer, bridge)
	// 		},
	// 	},
	// 	{
	// 		name: "Workspace Diagnostics Tool",
	// 		toolRegistration: func(mcpServer *server.MCPServer, bridge interfaces.BridgeInterface) {
	// 			tools.RegisterWorkspaceDiagnosticsTool(mcpServer, bridge)
	// 		},
	// 	},
	// }
	//
	// // Run each test case
	// for _, tc := range testCases {
	// 	t.Run(tc.name, func(t *testing.T) {
	// 		// Prepare mock bridge
	// 		mockBridge := &mocks.MockBridge{}
	//
	// 		// Set up expectations
	// 		// mockBridge.On("InferLanguage", "/test/example.go").Return("go", nil)
	// 		// mockBridge.On("InferLanguage", "/test/file.go").Return("go", nil)
	// 		// mockBridge.On("InferLanguage", mock.AnythingOfType("string")).Return("unknown", nil)
	// 		//
	// 		// mockBridge.On("GetClientForLanguageInterface", "go").Return(&lsp.LanguageClient{}, nil)
	// 		// mockBridge.On("GetClientForLanguageInterface", mock.AnythingOfType("string")).Return(nil, nil)
	// 		//
	// 		// mockBridge.On("GetConfig").Return(&lsp.LSPServerConfig{
	// 		// 	LanguageServers: map[string]lsp.LanguageServerConfig{
	// 		// 		"go": {Command: "gopls"},
	// 		// 	},
	// 		// })
	//
	// 		// mockBridge.On("CloseAllClients").Return()
	//
	// 		// Create MCP server
	// 		mcpServer := server.NewMCPServer(
	// 			"lsp-bridge-mcp",
	// 			"1.0.0",
	// 			server.WithToolCapabilities(false),
	// 		)
	//
	// 		// Register the tool - test passes if it doesn't panic
	// 		tc.toolRegistration(mcpServer, mockBridge)
	//
	// 		// Verify expectations were met
	// 		mockBridge.AssertExpectations(t)
	// 	})
	// }
}

// Benchmark tool registration performance
func BenchmarkMCPServerToolRegistration(b *testing.B) {
	mockBridge := &mocks.MockBridge{}

	for b.Loop() {
		mcpServer := server.NewMCPServer(
			"lsp-bridge-mcp",
			"1.0.0",
			server.WithToolCapabilities(false),
		)
		RegisterAllTools(mcpServer, mockBridge)
	}
}
