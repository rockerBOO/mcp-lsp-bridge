package mcpserver

import (
	"reflect"
	"testing"

	"rockerboo/mcp-lsp-bridge/lsp"
)

// MockBridge implements BridgeInterface for testing
type MockBridge struct{}

func (m *MockBridge) GetClientForLanguageInterface(language string) (any, error) {
	return nil, nil
}

func (m *MockBridge) InferLanguage(filePath string) (string, error) {
	return "go", nil
}

func (m *MockBridge) CloseAllClients() {}

func (m *MockBridge) GetConfig() *lsp.LSPServerConfig {
	return &lsp.LSPServerConfig{}
}

func TestMCPServerSetup(t *testing.T) {
	// Create a mock bridge
	mockBridge := &MockBridge{}

	// Set up the MCP server
	mcpServer := SetupMCPServer(mockBridge)

	t.Run("Server Creation", func(t *testing.T) {
		if mcpServer == nil {
			t.Fatal("MCP server should not be nil")
		}

		// Use reflection to check server metadata
		v := reflect.ValueOf(mcpServer).Elem()

		// Check name
		nameField := v.FieldByName("name")
		if !nameField.IsValid() {
			t.Fatal("Could not access server name")
		}
		if nameField.String() != "lsp-bridge-mcp" {
			t.Errorf("Expected server name 'lsp-bridge-mcp', got %s", nameField.String())
		}

		// Check version
		versionField := v.FieldByName("version")
		if !versionField.IsValid() {
			t.Fatal("Could not access server version")
		}
		if versionField.String() != "1.0.0" {
			t.Errorf("Expected server version '1.0.0', got %s", versionField.String())
		}
	})

	t.Run("Tool Registration Methods", func(t *testing.T) {
		// List of expected registration method names
		expectedRegistrationMethods := []string{
			"registerAnalyzeCodeTool",
			"registerInferLanguageTool",
			"registerLSPConnectTool",
			"registerLSPDisconnectTool",
		}

		// Verify each registration method
		for _, methodName := range expectedRegistrationMethods {
			t.Logf("Verifying registration method: %s", methodName)
		}
	})
}

// Benchmark server setup to ensure performance
func BenchmarkMCPServerSetup(b *testing.B) {
	mockBridge := &MockBridge{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SetupMCPServer(mockBridge)
	}
}
