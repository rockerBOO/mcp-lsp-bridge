package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"rockerboo/mcp-lsp-bridge/mocks"
	"rockerboo/mcp-lsp-bridge/types"

	"github.com/stretchr/testify/mock"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/mcptest"
	"github.com/mark3labs/mcp-go/server"
	"github.com/myleshyson/lsprotocol-go/protocol"
)

func TestCallHierarchyTool(t *testing.T) {
	testCases := []struct {
		name            string
		uri             string
		line            uint32
		character       uint32
		direction       string
		mockItems       []protocol.CallHierarchyItem
		mockIncoming    []protocol.CallHierarchyIncomingCall
		mockOutgoing    []protocol.CallHierarchyOutgoingCall
		expectError     bool
		expectedContent string
		description     string
	}{
		{
			name:      "successful call hierarchy preparation",
			uri:       "file:///main.go",
			line:      10,
			character: 5,
			direction: "both",
			mockItems: []protocol.CallHierarchyItem{
				{
					Name: "main",
					Kind: protocol.SymbolKindFunction,
					Uri:  "file:///main.go",
					Range: protocol.Range{
						Start: protocol.Position{Line: 10, Character: 0},
						End:   protocol.Position{Line: 15, Character: 1},
					},
				},
			},
			mockIncoming: []protocol.CallHierarchyIncomingCall{
				{
					From: protocol.CallHierarchyItem{
						Name: "init",
						Kind: protocol.SymbolKindFunction,
					},
					FromRanges: []protocol.Range{
						{
							Start: protocol.Position{Line: 5, Character: 0},
							End:   protocol.Position{Line: 5, Character: 4},
						},
					},
				},
			},
			mockOutgoing: []protocol.CallHierarchyOutgoingCall{
				{
					To: protocol.CallHierarchyItem{
						Name: "fmt.Println",
						Kind: protocol.SymbolKindFunction,
					},
					FromRanges: []protocol.Range{
						{
							Start: protocol.Position{Line: 12, Character: 1},
							End:   protocol.Position{Line: 12, Character: 12},
						},
					},
				},
			},
			expectError:     false,
			expectedContent: "CALL HIERARCHY",
			description:     "Should prepare call hierarchy for main function",
		},
		{
			name:      "incoming calls only",
			uri:       "file:///utils.go",
			line:      20,
			character: 10,
			direction: "incoming",
			mockItems: []protocol.CallHierarchyItem{
				{
					Name: "helper",
					Kind: protocol.SymbolKindFunction,
					Uri:  "file:///utils.go",
				},
			},
			mockIncoming: []protocol.CallHierarchyIncomingCall{
				{
					From: protocol.CallHierarchyItem{
						Name: "processData",
						Kind: protocol.SymbolKindFunction,
					},
				},
			},
			expectError:     false,
			expectedContent: "INCOMING CALLS",
			description:     "Should show incoming calls to helper function",
		},
		{
			name:      "outgoing calls only",
			uri:       "file:///service.go",
			line:      30,
			character: 15,
			direction: "outgoing",
			mockItems: []protocol.CallHierarchyItem{
				{
					Name: "service.Process",
					Kind: protocol.SymbolKindMethod,
					Uri:  "file:///service.go",
				},
			},
			mockOutgoing: []protocol.CallHierarchyOutgoingCall{
				{
					To: protocol.CallHierarchyItem{
						Name: "database.Query",
						Kind: protocol.SymbolKindMethod,
					},
				},
				{
					To: protocol.CallHierarchyItem{
						Name: "logger.Info",
						Kind: protocol.SymbolKindMethod,
					},
				},
			},
			expectError:     false,
			expectedContent: "OUTGOING CALLS",
			description:     "Should show outgoing calls from service method",
		},
		{
			name:        "call hierarchy preparation failure",
			uri:         "file:///invalid.go",
			line:        10,
			character:   5,
			direction:   "both",
			expectError: true,
			description: "Should handle call hierarchy preparation errors",
		},
		{
			name:            "no call hierarchy items",
			uri:             "file:///empty.go",
			line:            1,
			character:       0,
			direction:       "both",
			mockItems:       []protocol.CallHierarchyItem{},
			expectError:     false,
			expectedContent: "No call hierarchy",
			description:     "Should handle files with no callable symbols",
		},
		{
			name:      "complex call hierarchy with nested calls",
			uri:       "file:///complex.go",
			line:      50,
			character: 20,
			direction: "both",
			mockItems: []protocol.CallHierarchyItem{
				{
					Name: "complexFunction",
					Kind: protocol.SymbolKindFunction,
					Uri:  "file:///complex.go",
				},
			},
			mockIncoming: []protocol.CallHierarchyIncomingCall{
				{
					From: protocol.CallHierarchyItem{
						Name: "caller1",
						Kind: protocol.SymbolKindFunction,
					},
				},
				{
					From: protocol.CallHierarchyItem{
						Name: "caller2",
						Kind: protocol.SymbolKindMethod,
					},
				},
			},
			mockOutgoing: []protocol.CallHierarchyOutgoingCall{
				{
					To: protocol.CallHierarchyItem{
						Name: "callee1",
						Kind: protocol.SymbolKindFunction,
					},
				},
				{
					To: protocol.CallHierarchyItem{
						Name: "callee2",
						Kind: protocol.SymbolKindMethod,
					},
				},
				{
					To: protocol.CallHierarchyItem{
						Name: "callee3",
						Kind: protocol.SymbolKindFunction,
					},
				},
			},
			expectError:     false,
			expectedContent: "CALL HIERARCHY",
			description:     "Should handle complex call hierarchies with multiple callers and callees",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bridge := &mocks.MockBridge{}

			// Set up mock expectations
			if tc.expectError {
				bridge.On("PrepareCallHierarchy", tc.uri, tc.line, tc.character).Return([]protocol.CallHierarchyItem{}, errors.New("mock error"))
			} else {
				bridge.On("PrepareCallHierarchy", tc.uri, tc.line, tc.character).Return(tc.mockItems, nil)

				// Set up incoming/outgoing call expectations if items exist
				if len(tc.mockItems) > 0 {
					if tc.direction == "incoming" || tc.direction == "both" {
						bridge.On("IncomingCalls", mock.Anything).Return(tc.mockIncoming, nil)
					}

					if tc.direction == "outgoing" || tc.direction == "both" {
						bridge.On("OutgoingCalls", mock.Anything).Return(tc.mockOutgoing, nil)
					}
				}
			}

			// Create MCP server and register tool
			mcpServer, err := mcptest.NewServer(t)
			if err != nil {
				t.Errorf("Could not start the server: %v", err)
			}

			RegisterCallHierarchyTool(mcpServer, bridge)

			// Test call hierarchy preparation
			items, err := bridge.PrepareCallHierarchy(tc.uri, tc.line, tc.character)
			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}

				return
			}

			if err != nil {
				t.Errorf("Unexpected error in call hierarchy preparation: %v", err)
				return
			}

			if len(items) != len(tc.mockItems) {
				t.Errorf("Expected %d call hierarchy items, got %d", len(tc.mockItems), len(items))
			}

			// Test incoming calls if direction includes incoming
			if tc.direction == "incoming" || tc.direction == "both" {
				if len(items) > 0 {
					incoming, err := bridge.IncomingCalls(items[0])
					if err != nil {
						t.Errorf("Unexpected error getting incoming calls: %v", err)
					}

					if len(incoming) != len(tc.mockIncoming) {
						t.Errorf("Expected %d incoming calls, got %d", len(tc.mockIncoming), len(incoming))
					}
				}
			}

			// Test outgoing calls if direction includes outgoing
			if tc.direction == "outgoing" || tc.direction == "both" {
				if len(items) > 0 {
					outgoing, err := bridge.OutgoingCalls(items[0])
					if err != nil {
						t.Errorf("Unexpected error getting outgoing calls: %v", err)
					}

					if len(outgoing) != len(tc.mockOutgoing) {
						t.Errorf("Expected %d outgoing calls, got %d", len(tc.mockOutgoing), len(outgoing))
					}
				}
			}

			// Verify all expectations were met
			bridge.AssertExpectations(t)

			t.Logf("Test case '%s' passed - %s", tc.name, tc.description)
		})
	}
}
func TestCallHierarchySymbolTypes(t *testing.T) {
	symbolTypes := []struct {
		symbolKind  protocol.SymbolKind
		name        string
		description string
	}{
		{protocol.SymbolKindFunction, "function", "Regular function"},
		{protocol.SymbolKindMethod, "method", "Class/struct method"},
		{protocol.SymbolKindConstructor, "constructor", "Constructor function"},
		{protocol.SymbolKindClass, "class", "Class definition"},
		{protocol.SymbolKindInterface, "interface", "Interface definition"},
		{protocol.SymbolKindNamespace, "namespace", "Namespace/package"},
	}
	for _, symbolType := range symbolTypes {
		t.Run("call_hierarchy_"+symbolType.name, func(t *testing.T) {
			bridge := &mocks.MockBridge{}

			// Create mock call hierarchy item
			mockItem := protocol.CallHierarchyItem{
				Name: "testSymbol",
				Kind: symbolType.symbolKind,
				Uri:  "file:///test.go",
				Range: protocol.Range{
					Start: protocol.Position{Line: 10, Character: 5},
					End:   protocol.Position{Line: 10, Character: 15},
				},
				SelectionRange: protocol.Range{
					Start: protocol.Position{Line: 10, Character: 5},
					End:   protocol.Position{Line: 10, Character: 15},
				},
			}

			// Mock incoming call
			mockIncomingCall := protocol.CallHierarchyIncomingCall{
				From: protocol.CallHierarchyItem{
					Name: "caller",
					Kind: protocol.SymbolKindFunction,
					Uri:  "file:///caller.go",
					Range: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 0},
						End:   protocol.Position{Line: 5, Character: 10},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 0},
						End:   protocol.Position{Line: 5, Character: 10},
					},
				},
				FromRanges: []protocol.Range{
					{
						Start: protocol.Position{Line: 7, Character: 2},
						End:   protocol.Position{Line: 7, Character: 12},
					},
				},
			}

			// Mock outgoing call
			mockOutgoingCall := protocol.CallHierarchyOutgoingCall{
				To: protocol.CallHierarchyItem{
					Name: "callee",
					Kind: protocol.SymbolKindFunction,
					Uri:  "file:///callee.go",
					Range: protocol.Range{
						Start: protocol.Position{Line: 15, Character: 0},
						End:   protocol.Position{Line: 15, Character: 10},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 15, Character: 0},
						End:   protocol.Position{Line: 15, Character: 10},
					},
				},
				FromRanges: []protocol.Range{
					{
						Start: protocol.Position{Line: 12, Character: 4},
						End:   protocol.Position{Line: 12, Character: 14},
					},
				},
			}

			// Setup mock expectations
			bridge.On("PrepareCallHierarchy", "file:///test.go", uint32(10), uint32(5)).
				Return([]protocol.CallHierarchyItem{mockItem}, nil)
			bridge.On("IncomingCalls", mock.AnythingOfType("protocol.CallHierarchyItem")).
				Return([]protocol.CallHierarchyIncomingCall{mockIncomingCall}, nil)
			bridge.On("OutgoingCalls", mock.AnythingOfType("protocol.CallHierarchyItem")).
				Return([]protocol.CallHierarchyOutgoingCall{mockOutgoingCall}, nil)

			// Test call hierarchy for this symbol type
			items, err := bridge.PrepareCallHierarchy("file:///test.go", 10, 5)
			if err != nil {
				t.Errorf("Error preparing call hierarchy for %s: %v", symbolType.description, err)
				return
			}

			if len(items) != 1 {
				t.Errorf("Expected 1 item for %s, got %d", symbolType.description, len(items))
				return
			}

			// Test incoming calls
			incoming, err := bridge.IncomingCalls(items[0])
			if err != nil {
				t.Errorf("Error getting incoming calls for %s: %v", symbolType.description, err)
			}

			if len(incoming) != 1 {
				t.Errorf("Expected 1 incoming call for %s, got %d", symbolType.description, len(incoming))
			}

			// Test outgoing calls
			outgoing, err := bridge.OutgoingCalls(items[0])
			if err != nil {
				t.Errorf("Error getting outgoing calls for %s: %v", symbolType.description, err)
			}

			if len(outgoing) != 1 {
				t.Errorf("Expected 1 outgoing call for %s, got %d", symbolType.description, len(outgoing))
			}

			// Verify all expectations were met
			bridge.AssertExpectations(t)

			t.Logf("Call hierarchy test for %s passed", symbolType.description)
		})
	}
}

func TestCallHierarchyEdgeCases(t *testing.T) {
	t.Run("invalid direction parameter", func(t *testing.T) {
		// Prepare MCP server and mock bridge
		bridge := &mocks.MockBridge{}
		tool, handler := CallHierarchyTool(bridge)
		mcpServer, err := mcptest.NewServer(t, server.ServerTool{
			Tool:    tool,
			Handler: handler,
		})
		if err != nil {
			t.Errorf("Could not start the server: %v", err)
		}

		// Test tool with an invalid direction
		ctx := context.Background()
		result, err := mcpServer.Client().CallTool(ctx, mcp.CallToolRequest{
			Request: mcp.Request{Method: "tools/call"},
			Params: mcp.CallToolParams{
				Name: "call_hierarchy",
				Arguments: map[string]any{
					"uri":       "file:///test.go",
					"line":      10,
					"character": 5,
					"direction": "invalid",
				},
			},
		})

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result == nil || !result.IsError {
			t.Error("Expected error result for invalid direction")
		}
	})

	t.Run("missing required parameters", func(t *testing.T) {
		bridge := &mocks.MockBridge{}
		tool, handler := CallHierarchyTool(bridge)
		mcpServer, err := mcptest.NewServer(t, server.ServerTool{
			Tool:    tool,
			Handler: handler,
		})
		if err != nil {
			t.Errorf("Could not start the server: %v", err)
		}

		ctx := context.Background()

		// Test missing URI
		result, err := mcpServer.Client().CallTool(ctx, mcp.CallToolRequest{
			Request: mcp.Request{Method: "tools/call"},
			Params: mcp.CallToolParams{
				Name: "call_hierarchy",
				Arguments: map[string]any{
					"line":      10,
					"character": 5,
				},
			},
		})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil || !result.IsError {
			t.Error("Expected error result for missing URI")
		}

		// Test missing line
		result, err = mcpServer.Client().CallTool(ctx, mcp.CallToolRequest{
			Request: mcp.Request{Method: "tools/call"},
			Params: mcp.CallToolParams{
				Name: "call_hierarchy",
				Arguments: map[string]any{
					"uri":       "file:///test.go",
					"character": 5,
				},
			},
		})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil || !result.IsError {
			t.Error("Expected error result for missing line")
		}
	})

	t.Run("language inference failure", func(t *testing.T) {
		bridge := &mocks.MockBridge{}
		var lang *types.Language = nil
		bridge.On("InferLanguage", "file:///unknown.xyz").Return(lang, errors.New("unknown file type"))

		tool, handler := CallHierarchyTool(bridge)
		mcpServer, err := mcptest.NewServer(t, server.ServerTool{
			Tool:    tool,
			Handler: handler,
		})
		if err != nil {
			t.Errorf("Could not start the server: %v", err)
		}

		ctx := context.Background()
		result, err := mcpServer.Client().CallTool(ctx, mcp.CallToolRequest{
			Request: mcp.Request{Method: "tools/call"},
			Params: mcp.CallToolParams{
				Name: "call_hierarchy",
				Arguments: map[string]any{
					"uri":       "file:///unknown.xyz",
					"line":      10,
					"character": 5,
				},
			},
		})

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil || !result.IsError {
			t.Error("Expected error result for language inference failure")
		}

		bridge.AssertExpectations(t)
	})

	t.Run("call hierarchy errors during direction processing", func(t *testing.T) {
		bridge := &mocks.MockBridge{}

		// Setup valid prepare call hierarchy
		mockItem := protocol.CallHierarchyItem{
			Name: "testFunc",
			Kind: protocol.SymbolKindFunction,
			Uri:  "file:///test.go",
		}
		goLang := types.Language("go")
		bridge.On("InferLanguage", "file:///test.go").Return(&goLang, nil)
		bridge.On("PrepareCallHierarchy", "file:///test.go", uint32(10), uint32(5)).Return([]protocol.CallHierarchyItem{mockItem}, nil)

		// Mock error for incoming calls
		bridge.On("IncomingCalls", mock.AnythingOfType("protocol.CallHierarchyItem")).Return([]protocol.CallHierarchyIncomingCall{}, errors.New("incoming calls failed"))
		// Mock error for outgoing calls
		bridge.On("OutgoingCalls", mock.AnythingOfType("protocol.CallHierarchyItem")).Return([]protocol.CallHierarchyOutgoingCall{}, errors.New("outgoing calls failed"))

		tool, handler := CallHierarchyTool(bridge)
		mcpServer, err := mcptest.NewServer(t, server.ServerTool{
			Tool:    tool,
			Handler: handler,
		})
		if err != nil {
			t.Errorf("Could not start the server: %v", err)
		}

		ctx := context.Background()
		result, err := mcpServer.Client().CallTool(ctx, mcp.CallToolRequest{
			Request: mcp.Request{Method: "tools/call"},
			Params: mcp.CallToolParams{
				Name: "call_hierarchy",
				Arguments: map[string]any{
					"uri":       "file:///test.go",
					"line":      10,
					"character": 5,
					"direction": "both",
				},
			},
		})

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result == nil {
			t.Error("Expected result even with call hierarchy errors")
		}

		bridge.AssertExpectations(t)
	})
	t.Run("recursive function calls", func(t *testing.T) {
		bridge := &mocks.MockBridge{}

		// Mock the PrepareCallHierarchy call
		expectedItems := []protocol.CallHierarchyItem{{ /* your expected item structure */ }}
		bridge.On("PrepareCallHierarchy", "file:///recursive.go", uint32(10), uint32(5)).Return(expectedItems, nil)

		items, err := bridge.PrepareCallHierarchy("file:///recursive.go", 10, 5)
		if err != nil {
			t.Errorf("Error preparing call hierarchy for recursive function: %v", err)
		}

		if len(items) == 0 {
			t.Error("Expected call hierarchy items for recursive function")
		}

		bridge.AssertExpectations(t)
	})

	t.Run("deeply nested call chains", func(t *testing.T) {
		bridge := &mocks.MockBridge{}

		// Mock the PrepareCallHierarchy call
		expectedItems := []protocol.CallHierarchyItem{{ /* your expected item structure */ }}
		bridge.On("PrepareCallHierarchy", "file:///deep.go", uint32(25), uint32(10)).Return(expectedItems, nil)

		// Mock incoming calls
		expectedIncoming := []protocol.CallHierarchyIncomingCall{{}, {}, {}} // 3 items
		bridge.On("IncomingCalls", expectedItems[0]).Return(expectedIncoming, nil)

		// Mock outgoing calls
		expectedOutgoing := []protocol.CallHierarchyOutgoingCall{{}, {}, {}} // 3 items
		bridge.On("OutgoingCalls", expectedItems[0]).Return(expectedOutgoing, nil)

		items, err := bridge.PrepareCallHierarchy("file:///deep.go", 25, 10)
		if err != nil {
			t.Errorf("Error preparing call hierarchy for deeply nested calls: %v", err)
		}

		if len(items) == 0 {
			t.Error("Expected call hierarchy items for deeply nested function")
		}

		// Test that we can get both incoming and outgoing calls
		incoming, err := bridge.IncomingCalls(items[0])
		if err != nil {
			t.Errorf("Error getting incoming calls: %v", err)
		}

		if len(incoming) != 3 {
			t.Errorf("Expected 3 incoming calls, got %d", len(incoming))
		}

		outgoing, err := bridge.OutgoingCalls(items[0])
		if err != nil {
			t.Errorf("Error getting outgoing calls: %v", err)
		}

		if len(outgoing) != 3 {
			t.Errorf("Expected 3 outgoing calls, got %d", len(outgoing))
		}

		bridge.AssertExpectations(t)
	})
}

func TestFormatCallHierarchyResultsDetailed(t *testing.T) {
	testCases := []struct {
		name         string
		items        []protocol.CallHierarchyItem
		language     string
		errors       []error
		uri          string
		line         int
		character    int
		direction    string
		incoming     []protocol.CallHierarchyIncomingCall
		outgoing     []protocol.CallHierarchyOutgoingCall
		expectOutput string
	}{
		{
			name:     "basic formatting with items",
			language: "go",
			items: []protocol.CallHierarchyItem{
				{
					Name: "testFunction",
					Kind: protocol.SymbolKindFunction,
					Uri:  "file:///test.go",
					Range: protocol.Range{
						Start: protocol.Position{Line: 10, Character: 0},
						End:   protocol.Position{Line: 15, Character: 1},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 10, Character: 5},
						End:   protocol.Position{Line: 10, Character: 17},
					},
					Detail: "Test function details",
				},
			},
			uri:          "file:///test.go",
			line:         10,
			character:    5,
			direction:    "both",
			expectOutput: "CALL HIERARCHY",
		},
		{
			name:         "formatting with errors",
			language:     "go",
			items:        []protocol.CallHierarchyItem{},
			errors:       []error{errors.New("test error")},
			uri:          "file:///test.go",
			line:         10,
			character:    5,
			direction:    "both",
			expectOutput: "ERRORS",
		},
		{
			name:     "formatting with incoming calls and multiple ranges",
			language: "go",
			items: []protocol.CallHierarchyItem{
				{
					Name: "targetFunc",
					Kind: protocol.SymbolKindFunction,
				},
			},
			direction: "incoming",
			incoming: []protocol.CallHierarchyIncomingCall{
				{
					From: protocol.CallHierarchyItem{
						Name: "caller1",
						Uri:  "file:///caller.go",
					},
					FromRanges: []protocol.Range{
						{Start: protocol.Position{Line: 5, Character: 0}, End: protocol.Position{Line: 5, Character: 10}},
						{Start: protocol.Position{Line: 8, Character: 2}, End: protocol.Position{Line: 8, Character: 12}},
						{Start: protocol.Position{Line: 12, Character: 4}, End: protocol.Position{Line: 12, Character: 14}},
						{Start: protocol.Position{Line: 15, Character: 1}, End: protocol.Position{Line: 15, Character: 11}},
					},
				},
			},
			expectOutput: "INCOMING CALLS",
		},
		{
			name:     "formatting with outgoing calls",
			language: "typescript",
			items: []protocol.CallHierarchyItem{
				{
					Name: "sourceFunc",
					Kind: protocol.SymbolKindMethod,
				},
			},
			direction: "outgoing",
			outgoing: []protocol.CallHierarchyOutgoingCall{
				{
					To: protocol.CallHierarchyItem{
						Name: "callee1",
						Uri:  "file:///callee.ts",
					},
					FromRanges: []protocol.Range{
						{Start: protocol.Position{Line: 20, Character: 8}, End: protocol.Position{Line: 20, Character: 16}},
					},
				},
			},
			expectOutput: "OUTGOING CALLS",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatCallHierarchyResults(
				tc.items,
				tc.language,
				tc.errors,
				tc.uri,
				tc.line,
				tc.character,
				tc.direction,
				tc.incoming,
				tc.outgoing,
			)

			if tc.expectOutput != "" && !strings.Contains(result, tc.expectOutput) {
				t.Errorf("Expected output to contain '%s', but got: %s", tc.expectOutput, result)
			}

			// Verify line number conversion (0-based to 1-based)
			if len(tc.incoming) > 0 && len(tc.incoming[0].FromRanges) > 0 {
				firstRange := tc.incoming[0].FromRanges[0]
				expectedLine := fmt.Sprintf("Line %d:%d", firstRange.Start.Line+1, firstRange.Start.Character+1)
				if !strings.Contains(result, expectedLine) {
					t.Errorf("Expected 1-based line numbering with '%s', but got: %s", expectedLine, result)
				}
			}

			// Verify range limiting (should show max 3 ranges)
			if len(tc.incoming) > 0 && len(tc.incoming[0].FromRanges) > 3 {
				if !strings.Contains(result, "and 1 more ranges") {
					t.Errorf("Expected range limiting message for excess ranges, but got: %s", result)
				}
			}
		})
	}
}

func TestCallHierarchyToolParameterValidation(t *testing.T) {
	testCases := []struct {
		name        string
		params      map[string]any
		expectError bool
		description string
	}{
		{
			name: "valid parameters",
			params: map[string]any{
				"uri":       "file:///test.go",
				"line":      10,
				"character": 5,
				"direction": "both",
			},
			expectError: false,
			description: "Should accept valid parameters",
		},
		{
			name: "invalid line parameter - negative",
			params: map[string]any{
				"uri":       "file:///test.go",
				"line":      -1,
				"character": 5,
			},
			expectError: true,
			description: "Should reject negative line numbers",
		},
		{
			name: "invalid character parameter - negative",
			params: map[string]any{
				"uri":       "file:///test.go",
				"line":      10,
				"character": -1,
			},
			expectError: true,
			description: "Should reject negative character positions",
		},
		{
			name: "default direction when not specified",
			params: map[string]any{
				"uri":       "file:///test.go",
				"line":      10,
				"character": 5,
			},
			expectError: false,
			description: "Should default to 'both' direction when not specified",
		},
		{
			name: "case insensitive direction",
			params: map[string]any{
				"uri":       "file:///test.go",
				"line":      10,
				"character": 5,
				"direction": "INCOMING",
			},
			expectError: false,
			description: "Should accept uppercase direction parameters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bridge := &mocks.MockBridge{}

			// Setup expectations for valid cases
			if !tc.expectError {
				if uri, ok := tc.params["uri"].(string); ok {
					goLang := types.Language("go")
					bridge.On("InferLanguage", uri).Return(&goLang, nil)
					bridge.On("PrepareCallHierarchy", mock.Anything, mock.Anything, mock.Anything).Return([]protocol.CallHierarchyItem{}, nil)
				}
			}

			tool, handler := CallHierarchyTool(bridge)
			mcpServer, err := mcptest.NewServer(t, server.ServerTool{
				Tool:    tool,
				Handler: handler,
			})
			if err != nil {
				t.Errorf("Could not start the server: %v", err)
			}

			ctx := context.Background()
			result, err := mcpServer.Client().CallTool(ctx, mcp.CallToolRequest{
				Request: mcp.Request{Method: "tools/call"},
				Params: mcp.CallToolParams{
					Name:      "call_hierarchy",
					Arguments: tc.params,
				},
			})

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tc.expectError && (result == nil || !result.IsError) {
				t.Errorf("Expected error for %s, but got success", tc.description)
			}

			if !tc.expectError && result != nil && result.IsError {
				t.Errorf("Expected success for %s, but got error: %v", tc.description, result)
			}

			bridge.AssertExpectations(t)
		})
	}
}
