package tools

import (
	"context"
	"fmt"
	"strings"

	"rockerboo/mcp-lsp-bridge/async"
	"rockerboo/mcp-lsp-bridge/collections"
	"rockerboo/mcp-lsp-bridge/interfaces"
	"rockerboo/mcp-lsp-bridge/logger"
	"rockerboo/mcp-lsp-bridge/types"
	"rockerboo/mcp-lsp-bridge/utils"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/myleshyson/lsprotocol-go/protocol"
)

// RegisterCallHierarchyTool registers the call hierarchy tool
func RegisterCallHierarchyTool(mcpServer ToolServer, bridge interfaces.BridgeInterface) {
	mcpServer.AddTool(CallHierarchyTool(bridge))
}

func CallHierarchyTool(bridge interfaces.BridgeInterface) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("call_hierarchy",
			mcp.WithDescription("Show call hierarchy (callers and callees) for a symbol"),
			mcp.WithDestructiveHintAnnotation(false),
			mcp.WithString("uri", mcp.Description("URI to the file")),
			mcp.WithNumber("line", mcp.Description("Line number (0-based)")),
			mcp.WithNumber("character", mcp.Description("Character position (0-based)")),
			mcp.WithString("direction", mcp.Description("Direction: 'incoming', 'outgoing', or 'both' (default: 'both')")),
		), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Parse and validate parameters
			uri, err := request.RequireString("uri")
			if err != nil {
				logger.Error("call_hierarchy: URI parsing failed", err)
				return mcp.NewToolResultError(err.Error()), nil
			}

			line, err := request.RequireInt("line")
			if err != nil {
				logger.Error("call_hierarchy: Line parsing failed", err)
				return mcp.NewToolResultError(err.Error()), nil
			}

			character, err := request.RequireInt("character")
			if err != nil {
				logger.Error("call_hierarchy: Character parsing failed", err)
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Optional direction parameter - for future use
			_ = "both" // Default direction (unused for now)
			if direction, err := request.RequireString("direction"); err == nil {
				// Direction parameter exists but not used in current implementation
				_ = direction // TODO: Use direction when implementing call hierarchy filtering
			}

			// Validate parameters
			lineUint32, err := safeUint32(line)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid line number: %v", err)), nil
			}
			characterUint32, err := safeUint32(character)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid character position: %v", err)), nil
			}

			// Normalize URI to ensure proper file:// scheme
			normalizedURI := utils.NormalizeURI(uri)

			// Infer language from the specific file URI
			language, err := bridge.InferLanguage(normalizedURI)
			if err != nil {
				logger.Error("call_hierarchy: language inference failed", err)
				return mcp.NewToolResultError(fmt.Sprintf("Failed to infer language from URI: %v", err)), nil
			}

			// Get the specific language client for this file
			client, err := bridge.GetClientForLanguage(string(*language))
			if err != nil {
				logger.Error("call_hierarchy: failed to get language client", err)
				return mcp.NewToolResultError(fmt.Sprintf("No LSP client available for language %s", *language)), nil
			}

			// For call hierarchy, we primarily use the file's specific language client
			// But we could extend this to search across related languages in the future
			clients := map[types.Language]types.LanguageClientInterface{
				*language: client,
			}

			// Convert clients to async operations
			ops := collections.TransformMap(clients, func(client types.LanguageClientInterface) func() ([]protocol.CallHierarchyItem, error) {
				return func() ([]protocol.CallHierarchyItem, error) {
					return client.PrepareCallHierarchy(normalizedURI, lineUint32, characterUint32)
				}
			})

			// Execute call hierarchy preparation for the specific language
			results, err := async.MapWithKeys(ctx, ops)
			if err != nil {
				logger.Error("call_hierarchy: async execution failed", err)
				return mcp.NewToolResultError(fmt.Sprintf("Failed to execute call hierarchy preparation: %v", err)), nil
			}

			// Process results (should only be one result for the specific language)
			var allItems []protocol.CallHierarchyItem
			var errors []error
			var successfulLanguage string

			for _, result := range results {
				if result.Error != nil {
					errors = append(errors, fmt.Errorf("language %s: %w", result.Key, result.Error))
					logger.Error("call_hierarchy: language server error", fmt.Errorf("language %s: %w", result.Key, result.Error))
				} else {
					allItems = append(allItems, result.Value...)
					successfulLanguage = string(result.Key)
				}
			}

			if len(allItems) == 0 {
				var errorMsg strings.Builder
				fmt.Fprintf(&errorMsg, "No call hierarchy items found on line %d, character %d for language %s\n", line, character, *language)
				if len(errors) > 0 {
					fmt.Fprintf(&errorMsg, "\nError: %v\n", errors[0])
				}
				return mcp.NewToolResultText(errorMsg.String()), nil
			}

			// Format results
			result := formatCallHierarchyResults(allItems, successfulLanguage, errors, uri, line, character)

			return mcp.NewToolResultText(result), nil
		}
}

// formatCallHierarchyResults formats call hierarchy results for user-friendly output
func formatCallHierarchyResults(items []protocol.CallHierarchyItem, successfulLanguage string, errors []error, uri string, line, character int) string {
	var result strings.Builder

	// Header with summary
	fmt.Fprintf(&result, "=== CALL HIERARCHY ===\n")
	fmt.Fprintf(&result, "Position: %s:%d:%d\n", uri, line, character)
	fmt.Fprintf(&result, "Language: %s\n", successfulLanguage)
	fmt.Fprintf(&result, "Items found: %d\n", len(items))
	if len(errors) > 0 {
		fmt.Fprintf(&result, "Errors: %d\n", len(errors))
	}
	fmt.Fprintf(&result, "\n")

	// Show errors if any
	if len(errors) > 0 {
		fmt.Fprintf(&result, "=== ERRORS ===\n")
		for i, err := range errors {
			fmt.Fprintf(&result, "%d. %v\n", i+1, err)
		}
		fmt.Fprintf(&result, "\n")
	}

	// Show call hierarchy items
	if len(items) > 0 {
		fmt.Fprintf(&result, "=== CALL HIERARCHY ITEMS ===\n")
		for i, item := range items {
			fmt.Fprintf(&result, "%d. %s\n", i+1, item.Name)
			fmt.Fprintf(&result, "   Kind: %s\n", symbolKindToString(item.Kind))
			fmt.Fprintf(&result, "   URI: %s\n", item.Uri)
			fmt.Fprintf(&result, "   Range: %d:%d-%d:%d\n", 
				item.Range.Start.Line, item.Range.Start.Character,
				item.Range.End.Line, item.Range.End.Character)
			fmt.Fprintf(&result, "   Selection Range: %d:%d-%d:%d\n", 
				item.SelectionRange.Start.Line, item.SelectionRange.Start.Character,
				item.SelectionRange.End.Line, item.SelectionRange.End.Character)
			if item.Detail != "" {
				fmt.Fprintf(&result, "   Detail: %s\n", item.Detail)
			}
			if len(item.Tags) > 0 {
				fmt.Fprintf(&result, "   Tags: %v\n", item.Tags)
			}
			fmt.Fprintf(&result, "\n")
		}
	}

	fmt.Fprintf(&result, "Note: Full incoming/outgoing call analysis requires implementation in the bridge layer.\n")
	fmt.Fprintf(&result, "The above items can be used to query for incoming/outgoing calls when that feature is complete.\n")

	return result.String()
}
