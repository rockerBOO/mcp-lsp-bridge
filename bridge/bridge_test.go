package bridge

import (
	"path/filepath"
	"testing"

	"rockerboo/mcp-lsp-bridge/lsp"
)

// Mock configuration for testing
func createTestLSPConfig() *lsp.LSPServerConfig {
	return &lsp.LSPServerConfig{
		Global: struct {
			LogPath            string `json:"log_file_path"`
			LogLevel           string `json:"log_level"`
			MaxLogFiles        int    `json:"max_log_files"`
			MaxRestartAttempts int    `json:"max_restart_attempts"`
			RestartDelayMs     int    `json:"restart_delay_ms"`
		}{
			LogPath:     "/tmp/test.log",
			LogLevel:    "debug",
			MaxLogFiles: 5,
		},
		LanguageServers: map[string]lsp.LanguageServerConfig{
			"go": {
				Command:   "gopls",
				Args:      []string{""},
				Languages: []string{"go"},
				Filetypes: []string{".go"},
			},
			"mock": {
				Command:   "mock-lsp-server",
				Args:      []string{},
				Languages: []string{"mock"},
				Filetypes: []string{},
			},
		},
		ExtensionLanguageMap: map[string]string{
			".go": "go",
		},
	}
}

func TestLanguageClientConnection(t *testing.T) {
	// Create test bridge
	bridge := MCPLSPBridge{
		clients: make(map[string]*lsp.LanguageClient),
		config:  createTestLSPConfig(),
	}

	// Test echo client connection
	client, err := bridge.GetClientForLanguage("mock")
	if err != nil {
		t.Fatalf("Failed to get echo client: %v", err)
	}

	if client == nil {
		t.Fatal("Echo client is nil")
	}

	// Verify client is stored
	storedClient, exists := bridge.clients["mock"]
	if !exists {
		t.Fatal("Gopls client not stored in bridge")
	}

	if storedClient != client {
		t.Fatal("Stored client does not match returned client")
	}
}

func TestClientCaching(t *testing.T) {
	bridge := MCPLSPBridge{
		clients: make(map[string]*lsp.LanguageClient),
		config:  createTestLSPConfig(),
	}

	// First connection
	firstClient, err := bridge.GetClientForLanguage("mock")
	if err != nil {
		t.Fatalf("First connection failed: %v", err)
	}

	// Second connection should return cached client
	secondClient, err := bridge.GetClientForLanguage("mock")
	if err != nil {
		t.Fatalf("Second connection failed: %v", err)
	}

	if firstClient != secondClient {
		t.Fatal("Cached client not returned on second connection")
	}
}

func TestInvalidLanguageConnection(t *testing.T) {
	bridge := MCPLSPBridge{
		clients: make(map[string]*lsp.LanguageClient),
		config:  createTestLSPConfig(),
	}

	// Attempt to connect to non-existent language
	_, err := bridge.GetClientForLanguage("nonexistent")
	if err == nil {
		t.Fatal("Expected error for non-existent language, got nil")
	}
}

func TestNormalizeURI(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already proper file URI",
			input:    "file:///home/user/test.go",
			expected: "file:///home/user/test.go",
		},
		{
			name:     "absolute path",
			input:    "/home/user/test.go",
			expected: "file:///home/user/test.go",
		},
		{
			name:     "relative path",
			input:    "test.go",
			expected: "file://" + mustAbs("test.go"),
		},
		{
			name:     "other scheme",
			input:    "http://example.com",
			expected: "http://example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizeURI(tc.input)
			if result != tc.expected {
				t.Errorf("normalizeURI(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func mustAbs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return abs
}
