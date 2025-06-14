package lsp

import (
	"testing"
	"time"

	"github.com/myleshyson/lsprotocol-go/protocol"
)

func TestNewLanguageClient(t *testing.T) {
	// Use echo as a simple mock command
	client, err := NewLanguageClient("mock-lsp-server")
	if err != nil {
		t.Fatalf("Failed to create language client: %v", err)
	}
	defer client.Close()

	// Check basic initialization
	if client == nil {
		t.Fatal("NewLanguageClient returned nil")
	}

	if !client.IsConnected() {
		t.Error("Client should be marked as connected")
	}
}

func TestLanguageClientMetrics(t *testing.T) {
	client, err := NewLanguageClient("mock-lsp-server")
	if err != nil {
		t.Fatalf("Failed to create language client: %v", err)
	}
	defer client.Close()

	// Perform some requests to generate metrics
	for range 5 {
		err := client.SendRequest(
			"test/method",
			map[string]any{"key": "value"},
			&map[string]any{},
			1*time.Second,
		)
		// Ignore errors since we're using echo
		_ = err
	}

	// Retrieve and check metrics
	metrics := client.GetMetrics()

	if metrics["total_requests"].(int64) != 5 {
		t.Errorf("Expected 5 total requests, got %v", metrics["total_requests"])
	}

	// Verify other metric properties
	if metrics["command"] != "mock-lsp-server" {
		t.Errorf("Unexpected command: %v", metrics["command"])
	}

	t.Logf("Metrics: %v", metrics)

	// Since we're sending invalid method requests, the client should be marked as disconnected
	if metrics["is_connected"] != false {
		t.Error("Client should be marked as disconnected after failed requests")
	}
}

func TestLanguageClientClose(t *testing.T) {
	client, err := NewLanguageClient("echo")
	if err != nil {
		t.Fatalf("Failed to create language client: %v", err)
	}

	// Close the client
	err = client.Close()
	if err != nil {
		t.Errorf("Close() returned an error: %v", err)
	}

	// Verify post-close state
	if client.IsConnected() {
		t.Error("Client should not be connected after Close()")
	}

	if client.status != StatusUninitialized {
		t.Errorf("Expected status Uninitialized after Close(), got %v", client.status)
	}
}

func TestSendRequestErrorHandling(t *testing.T) {
	client, err := NewLanguageClient("nonexistent_command")
	if err == nil {
		t.Fatal("Expected error when creating client with nonexistent command")
	}

	// Client should be nil when creation fails
	if client != nil {
		t.Error("Expected nil client when creation fails")
		client.Close() // Clean up if somehow not nil
	}
}

func TestClientCapabilitiesAndServerCapabilities(t *testing.T) {
	client, err := NewLanguageClient("mock-lsp-server")
	if err != nil {
		t.Fatalf("Failed to create language client: %v", err)
	}
	defer client.Close()

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
		client.Close()
	}
}

func BenchmarkSendRequest(b *testing.B) {
	client, err := NewLanguageClient("mock-lsp-server")
	if err != nil {
		b.Fatalf("Failed to create language client: %v", err)
	}
	defer client.Close()

	for b.Loop() {
		err := client.SendRequest(
			"test/method",
			map[string]any{"key": "value"},
			&map[string]any{},
			1*time.Second,
		)
		// Ignore errors since echo doesn't understand LSP protocol
		_ = err
	}
}
