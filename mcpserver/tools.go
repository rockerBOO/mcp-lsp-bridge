package mcpserver

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"

	"rockerboo/mcp-lsp-bridge/logger"
	"rockerboo/mcp-lsp-bridge/lsp"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerAnalyzeCodeTool registers the analyze_code tool
func registerAnalyzeCodeTool(mcpServer *server.MCPServer, bridge BridgeInterface) {
	mcpServer.AddTool(mcp.NewTool("analyze_code",
		mcp.WithDescription("Analyze code for completion suggestions and insights"),
		mcp.WithString("uri", mcp.Description("URI to the file location to analyze")),
		mcp.WithNumber("line", mcp.Description("Line of the file to analyze")),
		mcp.WithNumber("character", mcp.Description("Character of the line to analyze")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		uri, err := request.RequireString("uri")
		if err != nil {
			logger.Error("analyze_code: URI parsing failed", err)
			return mcp.NewToolResultError(err.Error()), nil
		}

		line, err := request.RequireInt("line")
		if err != nil {
			logger.Error("analyze_code: Line parsing failed", err)
			return mcp.NewToolResultError(err.Error()), nil
		}

		character, err := request.RequireInt("character")
		if err != nil {
			logger.Error("analyze_code: Character parsing failed", err)
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Infer language from the file URI
		language, err := bridge.InferLanguage(uri)
		if err != nil {
			logger.Error("analyze_code: Language inference failed", err)
			return mcp.NewToolResultError("Could not infer language"), nil
		}

		// Get LSP client for the language
		client, err := bridge.GetClientForLanguageInterface(language)
		if err != nil {
			logger.Error("analyze_code: Failed to get LSP client", err)
			return mcp.NewToolResultError("Failed to get LSP client"), nil
		}

		if client == nil {
			logger.Error("analyze_code: Failed to get LSP client", err)
			return mcp.NewToolResultError("Failed to get LSP client"), nil
		}

		// Convert URI and cast client
		lspClient, ok := client.(*lsp.LanguageClient)
		if !ok {
			logger.Error("analyze_code: Invalid LSP client type")
			return mcp.NewToolResultError("Invalid LSP client type"), nil
		}

		// Perform code analysis
		analyzeOpts := lsp.AnalyzeCodeOptions{
			Uri:        uri,
			Line:       int32(line),
			Character:  int32(character),
			LanguageId: language,
		}

		result, err := lsp.AnalyzeCode(lspClient, analyzeOpts)
		if err != nil {
			logger.Error("analyze_code: Code analysis failed", err)
			return mcp.NewToolResultError("Code analysis failed"), nil
		}

		// Log analysis details
		logger.Info("analyze_code: Successfully analyzed code", 
			fmt.Sprintf("URI: %s, Line: %d, Character: %d", uri, line, character),
		)

		// Count completion suggestions by checking the length of the CompletionResponse
		completionCount := 0
		if result.Completion != nil {
			// Use reflection to handle different CompletionResponse types
			completionValue := reflect.ValueOf(result.Completion)
			if completionValue.Kind() == reflect.Ptr {
				completionValue = completionValue.Elem()
			}
			
			// Try to get the items or suggestions
			itemsField := completionValue.FieldByName("Items")
			if itemsField.IsValid() {
				completionCount = int(itemsField.Len())
			}
		}

		// Prepare result summary
		summary := fmt.Sprintf(
			"Analysis Results:\n" +
			"Hover: %v\n" +
			"Completion Suggestions: %d\n" +
			"Signature Help: %v\n" +
			"Diagnostics: %d\n" +
			"Code Actions: %d",
			result.Hover != nil,
			completionCount,
			result.SignatureHelp != nil,
			len(result.Diagnostics),
			len(result.CodeActions),
		)

		return mcp.NewToolResultText(summary), nil
	})
}

// registerInferLanguageTool registers the infer_language tool
func registerInferLanguageTool(mcpServer *server.MCPServer, bridge BridgeInterface) {
	mcpServer.AddTool(mcp.NewTool("infer_language",
		mcp.WithDescription("Infer the programming language for a file"),
		mcp.WithString("file_path", mcp.Description("Path to the file to infer language")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filePath, err := request.RequireString("file_path")
		if err != nil {
			logger.Error("infer_language: File path parsing failed", err)
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Infer language from file extension
		language, err := bridge.InferLanguage(filePath)
		if err != nil {
			ext := filepath.Ext(filePath)
			logger.Error("infer_language: Language inference failed", 
				fmt.Sprintf("Extension: %s", ext),
			)
			return mcp.NewToolResultError(fmt.Sprintf("No language found for extension %s", ext)), nil
		}

		logger.Info("infer_language: Successfully inferred language", 
			fmt.Sprintf("File: %s, Language: %s", filePath, language),
		)

		return mcp.NewToolResultText(language), nil
	})
}

// registerLSPConnectTool registers the lsp_connect tool
func registerLSPConnectTool(mcpServer *server.MCPServer, bridge BridgeInterface) {
	mcpServer.AddTool(mcp.NewTool("lsp_connect",
		mcp.WithDescription("Connect to a language server for a specific language"),
		mcp.WithString("language", mcp.Description("Programming language to connect")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		language, err := request.RequireString("language")
		if err != nil {
			logger.Error("lsp_connect: Language parsing failed", err)
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Check if language server is configured
		config := bridge.GetConfig()
		if config == nil {
			logger.Error("lsp_connect: No configuration available")
			return mcp.NewToolResultError("No configuration available"), nil
		}

		if _, exists := config.LanguageServers[language]; !exists {
			logger.Error("lsp_connect: No language server configured", 
				fmt.Sprintf("Language: %s", language),
			)
			return mcp.NewToolResultError(fmt.Sprintf("No language server configured for %s", language)), nil
		}

		// Attempt to get or create the LSP client
		_, err = bridge.GetClientForLanguageInterface(language)
		if err != nil {
			logger.Error("lsp_connect: Failed to set up LSP client", 
				fmt.Sprintf("Language: %s, Error: %v", language, err),
			)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to set up LSP client: %v", err)), nil
		}

		logger.Info("lsp_connect: Successfully connected to LSP", 
			fmt.Sprintf("Language: %s", language),
		)

		return mcp.NewToolResultText(fmt.Sprintf("Connected to LSP for %s", language)), nil
	})
}

// registerLSPDisconnectTool registers the lsp_disconnect tool
func registerLSPDisconnectTool(mcpServer *server.MCPServer, bridge BridgeInterface) {
	mcpServer.AddTool(mcp.NewTool("lsp_disconnect",
		mcp.WithDescription("Disconnect all active language server clients"),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Close all active clients
		bridge.CloseAllClients()

		logger.Info("lsp_disconnect: Disconnected all language server clients")

		return mcp.NewToolResultText("All language server clients disconnected"), nil
	})
}
