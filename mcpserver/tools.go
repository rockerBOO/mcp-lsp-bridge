package mcpserver

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"rockerboo/mcp-lsp-bridge/logger"
	"rockerboo/mcp-lsp-bridge/lsp"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/myleshyson/lsprotocol-go/protocol"
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
			"Analysis Results:\n"+
				"Hover: %v\n"+
				"Completion Suggestions: %d\n"+
				"Signature Help: %v\n"+
				"Diagnostics: %d\n"+
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

func registerProjectAnalysisTool(mcpServer *server.MCPServer, bridge BridgeInterface) {
	mcpServer.AddTool(mcp.NewTool("project_analysis",
		mcp.WithDescription("Analyze project structure, find references, and search across files"),
		mcp.WithString("workspace_uri", mcp.Description("URI to the workspace/project root")),
		mcp.WithString("query", mcp.Description("Symbol or text to search for")),
		mcp.WithString("analysis_type", mcp.Description("Type of analysis: 'references', 'definitions', 'workspace_symbols', or 'text_search'")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		workspaceUri, err := request.RequireString("workspace_uri")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		query, err := request.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		analysisType, err := request.RequireString("analysis_type")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Convert URI to local file path
		projectPath := strings.TrimPrefix(workspaceUri, "file://")

		// Use the project language detection method instead of single file inference
		languages, err := bridge.DetectProjectLanguages(projectPath)

		if err != nil {
			logger.Error("Project language detection failed", fmt.Sprintf("Workspace URI: %s, Error: %v", workspaceUri, err))
			return mcp.NewToolResultError(fmt.Sprintf("Failed to detect project languages: %v", err)), nil
		}

		// Use the first detected language
		if len(languages) == 0 {
			logger.Warn("No programming languages detected in project", fmt.Sprintf("Workspace URI: %s", workspaceUri))
			return mcp.NewToolResultError("No languages detected in project"), nil
		}

		// Try to get clients for multiple languages with fallback
		clients, err := bridge.GetMultiLanguageClients(languages)
		if err != nil || len(clients) == 0 {
			return mcp.NewToolResultError("No LSP clients available for detected languages"), nil
		}

		// Use the first available client
		var lspClient *lsp.LanguageClient
		var activeLanguage string
		for lang, client := range clients {
			if typedClient, ok := client.(*lsp.LanguageClient); ok {
				lspClient = typedClient
				activeLanguage = lang
				break
			}
		}

		if lspClient == nil {
			return mcp.NewToolResultError("Invalid LSP client type"), nil
		}

		var response strings.Builder
		response.WriteString(fmt.Sprintf("Project Analysis: %s\n", analysisType))
		response.WriteString(fmt.Sprintf("Query: %s\n", query))
		response.WriteString(fmt.Sprintf("Workspace: %s\n", workspaceUri))
		response.WriteString(fmt.Sprintf("Detected Languages: %v\n", languages))
		response.WriteString(fmt.Sprintf("Active Language: %s\n\n", activeLanguage))

		switch analysisType {
		case "workspace_symbols":
			symbols, err := lspClient.WorkspaceSymbols(query)
			if err != nil {
				return mcp.NewToolResultError("Failed to get workspace symbols"), nil
			}

			formatWorkspaceSymbols := func(symbols []protocol.SymbolInformation) string {
				var result strings.Builder
				for i, symbol := range symbols {
					result.WriteString(fmt.Sprintf("%d. %v\n", i+1, symbol))
				}
				return result.String()
			}

			response.WriteString("=== WORKSPACE SYMBOLS ===\n")
			response.WriteString(formatWorkspaceSymbols(symbols))

		case "references":
			// For references, we need to search for the symbol first
			symbols, err := lspClient.WorkspaceSymbols(query)
			if err != nil {
				return mcp.NewToolResultError("Failed to get workspace symbols for reference search"), nil
			}

			response.WriteString("=== REFERENCES ===\n")
			if len(symbols) == 0 {
				response.WriteString("No symbols found matching the query.\n")
				break
			}

			// Use the first symbol found to get references
			symbol := symbols[0]
			// Extract position from symbol location
			uri := string(symbol.Location.Uri)
			line := symbol.Location.Range.Start.Line
			character := symbol.Location.Range.Start.Character

			references, err := bridge.FindSymbolReferences(activeLanguage, uri, int32(line), int32(character), true)
			if err != nil {
				response.WriteString(fmt.Sprintf("Failed to find references: %v\n", err))
				break
			}

			for i, ref := range references {
				response.WriteString(fmt.Sprintf("%d. %v\n", i+1, ref))
			}

		case "definitions":
			// For definitions, search for the symbol first
			symbols, err := lspClient.WorkspaceSymbols(query)
			if err != nil {
				return mcp.NewToolResultError("Failed to get workspace symbols for definition search"), nil
			}

			response.WriteString("=== DEFINITIONS ===\n")
			if len(symbols) == 0 {
				response.WriteString("No symbols found matching the query.\n")
				break
			}

			// Use the first symbol found to get definitions
			symbol := symbols[0]
			uri := string(symbol.Location.Uri)
			line := symbol.Location.Range.Start.Line
			character := symbol.Location.Range.Start.Character

			definitions, err := bridge.FindSymbolDefinitions(activeLanguage, uri, int32(line), int32(character))
			if err != nil {
				response.WriteString(fmt.Sprintf("Failed to find definitions: %v\n", err))
				break
			}

			for i, def := range definitions {
				response.WriteString(fmt.Sprintf("%d. %v\n", i+1, def))
			}

		case "text_search":
			response.WriteString("=== TEXT SEARCH ===\n")
			// Use workspace symbols as a text search mechanism
			searchResults, err := bridge.SearchTextInWorkspace(activeLanguage, query)
			if err != nil {
				response.WriteString(fmt.Sprintf("Failed to perform text search: %v\n", err))
				break
			}

			if len(searchResults) == 0 {
				response.WriteString("No results found for the search query.\n")
				break
			}

			for i, result := range searchResults {
				response.WriteString(fmt.Sprintf("%d. %v\n", i+1, result))
			}

		default:
			return mcp.NewToolResultError(fmt.Sprintf("Unknown analysis type: %s", analysisType)), nil
		}

		return mcp.NewToolResultText(response.String()), nil
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
		config := bridge.GetConfig()
		if config == nil {
			logger.Error("infer_language: No configuration available")
			return mcp.NewToolResultError("No LSP configuration found"), nil
		}

		ext := filepath.Ext(filePath)
		language, found := config.ExtensionLanguageMap[ext]
		if !found {
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

// registerProjectLanguageDetectionTool registers the detect_project_languages tool
func registerProjectLanguageDetectionTool(mcpServer *server.MCPServer, bridge BridgeInterface) {
	mcpServer.AddTool(mcp.NewTool("detect_project_languages",
		mcp.WithDescription("Detect all programming languages used in a project by examining root markers and file extensions"),
		mcp.WithString("project_path", mcp.Description("Path to the project directory to analyze")),
		mcp.WithString("mode", mcp.Description("Detection mode: 'all' for all languages, 'primary' for primary language only (default: 'all')")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectPath, err := request.RequireString("project_path")
		if err != nil {
			logger.Error("detect_project_languages: Project path parsing failed", err)
			return mcp.NewToolResultError(err.Error()), nil
		}

		mode, err := request.RequireString("mode")
		if err != nil {
			// Default to "all" if mode is not specified
			mode = "all"
		}

		logger.Info("detect_project_languages: Starting language detection",
			fmt.Sprintf("Path: %s, Mode: %s", projectPath, mode),
		)

		switch mode {
		case "primary":
			primaryLanguage, err := bridge.DetectPrimaryProjectLanguage(projectPath)
			if err != nil {
				logger.Error("detect_project_languages: Primary language detection failed", err)
				return mcp.NewToolResultError(fmt.Sprintf("Failed to detect primary language: %v", err)), nil
			}

			logger.Info("detect_project_languages: Primary language detected",
				fmt.Sprintf("Language: %s", primaryLanguage),
			)

			return mcp.NewToolResultText(fmt.Sprintf("Primary language: %s", primaryLanguage)), nil

		case "all":
			fallthrough
		default:
			languages, err := bridge.DetectProjectLanguages(projectPath)
			if err != nil {
				logger.Error("detect_project_languages: Language detection failed", err)
				return mcp.NewToolResultError(fmt.Sprintf("Failed to detect languages: %v", err)), nil
			}

			if len(languages) == 0 {
				return mcp.NewToolResultText("No programming languages detected in project"), nil
			}

			logger.Info("detect_project_languages: Languages detected",
				fmt.Sprintf("Count: %d, Languages: %v", len(languages), languages),
			)

			// Format the result
			result := "Detected languages (in priority order):\n"
			for i, lang := range languages {
				priority := "Primary"
				if i > 0 {
					priority = "Secondary"
				}
				result += fmt.Sprintf("%d. %s (%s)\n", i+1, lang, priority)
			}

			return mcp.NewToolResultText(result), nil
		}
	})
}
