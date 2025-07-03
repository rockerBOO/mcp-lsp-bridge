package lsp

import (
	"testing"

	"github.com/myleshyson/lsprotocol-go/protocol"
)

func TestNewLanguageClient(t *testing.T) {
	// This is a unit test for client creation, not connection
	client, err := NewLanguageClient("mock-lsp-server")
	if err != nil {
		t.Fatalf("Failed to create language client: %v", err)
	}

	// Check basic initialization without connecting
	if client == nil {
		t.Fatal("NewLanguageClient returned nil")
	}

	// Initial status should be connecting
	if client.Status() != StatusConnecting {
		t.Errorf("Expected initial status to be StatusConnecting, got %v", client.Status())
	}

	// Not connected initially
	if client.IsConnected() {
		t.Error("Client should not be marked as connected before connecting")
	}

	// Should be able to close without connecting
	if err := client.Close(); err != nil {
		t.Errorf("Close() should not error on unconnected client: %v", err)
	}
}

func TestLanguageClientMetrics(t *testing.T) {
	// Test metrics without connecting - just test the metrics structure
	client, err := NewLanguageClient("mock-lsp-server")
	if err != nil {
		t.Fatalf("Failed to create language client: %v", err)
	}

	// Get initial metrics
	metrics := client.GetMetrics()

	// Verify initial metric properties
	if metrics.Command != "mock-lsp-server" {
		t.Errorf("Unexpected command: %v", metrics.Command)
	}

	if metrics.TotalRequests != 0 {
		t.Errorf("Expected 0 total requests initially, got %v", metrics.TotalRequests)
	}

	if metrics.SuccessfulRequests != 0 {
		t.Errorf("Expected 0 successful requests initially, got %v", metrics.SuccessfulRequests)
	}

	if metrics.FailedRequests != 0 {
		t.Errorf("Expected 0 failed requests initially, got %v", metrics.FailedRequests)
	}

	if metrics.IsConnected != false {
		t.Error("Client should not be marked as connected initially")
	}

	if metrics.Status != StatusConnecting {
		t.Errorf("Expected initial status StatusConnecting, got %v", metrics.Status)
	}
}

func TestLanguageClientClose(t *testing.T) {
	// Test close without connecting
	client, err := NewLanguageClient("echo")
	if err != nil {
		t.Fatalf("Failed to create language client: %v", err)
	}

	// Close the client without connecting
	err = client.Close()
	if err != nil {
		t.Errorf("Close() returned an error: %v", err)
	}

	// Verify post-close state
	if client.IsConnected() {
		t.Error("Client should not be connected after Close()")
	}

	if client.Status() != StatusUninitialized {
		t.Errorf("Expected status Uninitialized after Close(), got %v", client.Status())
	}
}

func TestSendRequestErrorHandling(t *testing.T) {
	client, err := NewLanguageClient("nonexistent_command")
	if err != nil {
		t.Fatal("Expected error when creating client with nonexistent command")
	}

	connected_client, err := client.Connect()
	if err == nil {
		t.Fatalf("Expected to error when connecting to client: %v", err)
	}

	if connected_client != nil {
		t.Error("Expected nil client when creation fails")
		// Clean up if somehow not nil
		if err := client.Close(); err != nil {
			t.Logf("failed to close client: %v", err)
		}
	}
}

func TestClientCapabilitiesAndServerCapabilities(t *testing.T) {
	// Test capabilities without connecting
	client, err := NewLanguageClient("mock-lsp-server")
	if err != nil {
		t.Fatalf("Failed to create language client: %v", err)
	}

	// Test client capabilities
	clientCaps := client.ClientCapabilities()
	if clientCaps != (protocol.ClientCapabilities{}) {
		t.Error("Initial client capabilities should be empty")
	}

	// Test server capabilities setting
	testServerCaps := protocol.ServerCapabilities{
		TextDocumentSync: &protocol.Or2[protocol.TextDocumentSyncOptions, protocol.TextDocumentSyncKind]{
			Value: protocol.TextDocumentSyncKind(1),
		},
	}
	client.SetServerCapabilities(testServerCaps)

	serverCaps := client.ServerCapabilities()
	if serverCaps.TextDocumentSync == nil {
		t.Error("Server capabilities not set correctly")
	}
}

// Benchmark client creation and basic operations
func BenchmarkLanguageClientCreation(b *testing.B) {
	for b.Loop() {
		client, err := NewLanguageClient("mock-lsp-server")
		if err != nil {
			b.Fatalf("Failed to create language client: %v", err)
		}

		if err := client.Close(); err != nil {
			b.Logf("failed to close client: %v", err)
		}
	}
}

func BenchmarkGetMetrics(b *testing.B) {
	client, err := NewLanguageClient("mock-lsp-server")
	if err != nil {
		b.Fatalf("Failed to create language client: %v", err)
	}

	defer func() {
		if err := client.Close(); err != nil {
			b.Logf("failed to close client: %v", err)
		}
	}()

	for b.Loop() {
		_ = client.GetMetrics()
	}
}

