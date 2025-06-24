package tools

import (
	"context"
	"fmt"

	"rockerboo/mcp-lsp-bridge/interfaces"
	"rockerboo/mcp-lsp-bridge/logger"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func HoverTool(bridge interfaces.BridgeInterface) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("hover",
			mcp.WithDescription("Get detailed symbol information (signatures, documentation, types) at precise coordinates. Position-sensitive - use 'project_analysis' with 'definitions' first to find exact coordinates."),
			mcp.WithString("uri", mcp.Description("URI to the file")),
			mcp.WithNumber("line", mcp.Description("Line number (0-based) - use coordinates from 'definitions' for best results")),
			mcp.WithNumber("character", mcp.Description("Character position (0-based) - target middle of symbol identifier")),
		), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Extensive debug logging
			logger.Info("Hover Tool: Starting hover information request")

			// Parse and validate parameters
			uri, err := request.RequireString("uri")
			if err != nil {
				logger.Error("hover: URI parsing failed", err)
				return mcp.NewToolResultError(err.Error()), nil
			}
			logger.Info(fmt.Sprintf("Hover Tool: URI parsed: %s", uri))

			line, err := request.RequireInt("line")
			if err != nil {
				logger.Error("hover: Line parsing failed", err)
				return mcp.NewToolResultError(err.Error()), nil
			}
			logger.Info(fmt.Sprintf("Hover Tool: Line parsed: %d", line))

			character, err := request.RequireInt("character")
			if err != nil {
				logger.Error("hover: Character parsing failed", err)
				return mcp.NewToolResultError(err.Error()), nil
			}
			logger.Info(fmt.Sprintf("Hover Tool: Character parsed: %d", character))

			// Infer language for debugging
			language, langErr := bridge.InferLanguage(uri)
			if langErr != nil {
				logger.Error("hover: Language inference failed", langErr)
			} else {
				logger.Info(fmt.Sprintf("Hover Tool: Inferred language: %s", language))
			}

			// Execute bridge method with detailed error logging
			result, err := bridge.GetHoverInformation(uri, uint32(line), uint32(character))
			if err != nil {
				logger.Error("hover: Request failed", fmt.Sprintf("URI: %s, Line: %d, Character: %d, Error: %v", uri, line, character, err))
				return mcp.NewToolResultError(fmt.Sprintf("Failed to get hover information: %v", err)), nil
			}

			// Detailed logging of result
			if result == nil {
				logger.Info("Hover Tool: No hover information available")
				return mcp.NewToolResultText("No hover information available"), nil
			}

			// Enhanced result type logging
			logger.Info(fmt.Sprintf("Hover Tool: Result type: %T", result))

			content := formatHoverContent(result.Contents)
			return mcp.NewToolResultText(content), nil
		}
}

// RegisterHoverTool registers the hover tool with the MCP server.
func RegisterHoverTool(mcpServer ToolServer, bridge interfaces.BridgeInterface) {
	mcpServer.AddTool(HoverTool(bridge))
}
