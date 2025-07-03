package lsp_test

import (
	"rockerboo/mcp-lsp-bridge/lsp"
	"testing"
	"time"

	"github.com/myleshyson/lsprotocol-go/protocol"
)

// [Previous test functions remain the same]

func TestLanguageClient_SendRequest(t *testing.T) {
	// This is a unit test for SendRequest without actual connection
	// Testing the SendRequest method logic, not the actual JSON-RPC communication
	
	t.Run("SendRequest without connection should return error", func(t *testing.T) {
		lc, err := lsp.NewLanguageClient("nonexistent-command")
		if err != nil {
			t.Fatalf("Failed to create language client: %v", err)
		}

		// Try to send request without connecting first
		params := protocol.InitializeParams{
			ProcessId: func() *int32 { v := int32(12345); return &v }(),
		}
		result := &protocol.InitializeResult{}
		
		err = lc.SendRequest("initialize", params, result, 1*time.Second)
		
		// Should fail because no connection was established
		if err == nil {
			t.Error("Expected SendRequest to fail without connection, but it succeeded")
		}
	})
	
	t.Run("SendRequest error handling", func(t *testing.T) {
		lc, err := lsp.NewLanguageClient("echo")
		if err != nil {
			t.Fatalf("Failed to create language client: %v", err)
		}

		// Test with invalid parameters (empty method)
		err = lc.SendRequest("", nil, nil, 0)
		if err == nil {
			t.Error("Expected SendRequest to fail with empty method name")
		}
	})
}
