package tools

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"rockerboo/mcp-lsp-bridge/async"
	"rockerboo/mcp-lsp-bridge/interfaces"
	"rockerboo/mcp-lsp-bridge/logger"
	"rockerboo/mcp-lsp-bridge/types"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/myleshyson/lsprotocol-go/protocol"
)

// SymbolMatch represents a symbol found during exploration
type SymbolMatch struct {
	Name           string
	Kind           protocol.SymbolKind
	Location       protocol.Location
	ContainerName  string
	Documentation  string
	Signature      string
	ReferenceCount int
	Preview        string
}

// SymbolSessionData stores session-specific symbol exploration state
type SymbolSessionData struct {
	LastQuery     string
	SearchResults []SymbolMatch
	ActiveSymbol  *SymbolMatch
}

func SymbolExploreTool(bridge interfaces.BridgeInterface) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("symbol_explore",
		mcp.WithDescription(`Explore code symbols with intelligent search and context-aware information gathering.

USAGE EXAMPLES:
• Find symbol: query="getUserData"
• Filter by file: query="getUserData", file_context="auth"
• Get details: query="connectDB", detail_level="full"

PARAMETERS:
• query: Symbol name to search for (required)
• file_context: Fuzzy file filter (optional, e.g., "auth", "utils", "models")
• detail_level: Information depth - "auto", "basic", "full" (default: "auto")
• workspace_scope: Search scope - "project", "current_dir" (default: "project")

BEHAVIOR:
• Single match: Returns full details immediately
• Multiple matches: Returns summary with disambiguation options
• Uses session state for progressive exploration`),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithString("query", mcp.Description("Symbol name to search for"), mcp.Required()),
		mcp.WithString("file_context", mcp.Description("Fuzzy file filter (filename, directory, or path component)")),
		mcp.WithString("detail_level", mcp.Description("Information depth: auto, basic, full")),
		mcp.WithString("workspace_scope", mcp.Description("Search scope: project, current_dir")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get session from context
		session := server.ClientSessionFromContext(ctx)
		if session == nil {
			return mcp.NewToolResultError("No active session"), nil
		}

		// Parse parameters
		query, err := request.RequireString("query")
		if err != nil {
			logger.Error("symbol_explore: Query parsing failed", err)
			return mcp.NewToolResultError(err.Error()), nil
		}

		fileContext := request.GetString("file_context", "")
		detailLevel := request.GetString("detail_level", "auto")

		logger.Info(fmt.Sprintf("Symbol Explore: query=%s, file_context=%s, detail_level=%s", query, fileContext, detailLevel))

		// Session validation - for now we'll proceed without session dependency  
		_ = server.ClientSessionFromContext(ctx) // We're not using sessions yet

		// For now, we'll use a simple session ID-based approach
		// TODO: Implement proper session data storage when MCP sessions are fully set up
		sessionData := &SymbolSessionData{
			SearchResults: make([]SymbolMatch, 0),
		}

		// Perform workspace symbol search across multiple languages asynchronously
		symbols, err := performSymbolSearch(ctx, bridge, query)
		if err != nil {
			logger.Error("symbol_explore: Workspace symbol search failed", err)
			return mcp.NewToolResultError(fmt.Sprintf("Symbol search failed: %v", err)), nil
		}

		if len(symbols) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No symbols found matching '%s'", query)), nil
		}

		// Filter by file context if provided
		filteredSymbols := symbols
		if fileContext != "" {
			filteredSymbols = filterSymbolsByFileContext(symbols, fileContext)
		}

		// Store results in session data
		sessionData.LastQuery = query
		sessionData.SearchResults = filteredSymbols

		// Generate response based on results and detail level
		return generateSymbolResponse(bridge, filteredSymbols, query, fileContext, detailLevel)
	}
}

// performSymbolSearch executes workspace symbol search across multiple languages asynchronously
func performSymbolSearch(ctx context.Context, bridge interfaces.BridgeInterface, query string) ([]SymbolMatch, error) {
	// TODO: Fix architecture - project directory should be stored in bridge at startup
	// rather than calling os.Getwd() here. This is a workaround.
	projectDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	// First detect all project languages from the project directory
	languages, err := bridge.DetectProjectLanguages(projectDir)
	if err != nil {
		return nil, fmt.Errorf("failed to detect project languages: %w", err)
	}

	if len(languages) == 0 {
		return nil, errors.New("no languages detected in project directory")
	}

	logger.Info(fmt.Sprintf("Searching symbols across %d languages: %v", len(languages), languages))

	// Create async operations for each language
	operations := make(map[types.Language]func() ([]protocol.WorkspaceSymbol, error))
	for _, lang := range languages {
		language := lang // Capture for closure
		operations[language] = func() ([]protocol.WorkspaceSymbol, error) {
			return bridge.SearchTextInWorkspace(string(language), query)
		}
	}

	// Execute searches asynchronously
	results, err := async.MapWithKeys(ctx, operations)
	if err != nil {
		return nil, fmt.Errorf("async symbol search failed: %w", err)
	}

	// Collect all symbols from successful searches
	allMatches := make([]SymbolMatch, 0)
	for _, result := range results {
		if result.Error != nil {
			logger.Warn(fmt.Sprintf("Symbol search failed for language %s", result.Key), result.Error)
			continue
		}

		// Convert workspace symbols to our SymbolMatch format
		for _, symbol := range result.Value {
			match := convertWorkspaceSymbolToMatch(symbol)
			allMatches = append(allMatches, match)
		}
	}

	logger.Info(fmt.Sprintf("Found %d total symbols across all languages", len(allMatches)))
	return allMatches, nil
}

// convertWorkspaceSymbolToMatch converts a protocol.WorkspaceSymbol to our SymbolMatch type
func convertWorkspaceSymbolToMatch(symbol protocol.WorkspaceSymbol) SymbolMatch {
	// Handle the Or2[Location, LocationUriOnly] type
	var location protocol.Location
	switch v := symbol.Location.Value.(type) {
	case protocol.Location:
		location = v
	case protocol.LocationUriOnly:
		// For LocationUriOnly, create a Location with empty range
		location = protocol.Location{
			Uri:   v.Uri,
			Range: protocol.Range{}, // Empty range
		}
	default:
		// Fallback - create empty location
		location = protocol.Location{
			Uri:   "",
			Range: protocol.Range{},
		}
	}

	return SymbolMatch{
		Name:          symbol.Name,
		Kind:          symbol.Kind,
		Location:      location,
		ContainerName: symbol.ContainerName,
	}
}


// filterSymbolsByFileContext applies fuzzy file filtering
func filterSymbolsByFileContext(symbols []SymbolMatch, fileContext string) []SymbolMatch {
	if fileContext == "" {
		return symbols
	}

	filtered := make([]SymbolMatch, 0)
	fileContext = strings.ToLower(fileContext)

	for _, symbol := range symbols {
		uri := string(symbol.Location.Uri)
		fileName := filepath.Base(uri)
		dirName := filepath.Dir(uri)
		
		// Score the match
		score := 0
		
		// Exact filename match (highest score)
		if strings.Contains(strings.ToLower(fileName), fileContext) {
			score += 100
		}
		
		// Directory name match
		if strings.Contains(strings.ToLower(dirName), fileContext) {
			score += 50
		}
		
		// Path component match
		pathParts := strings.Split(strings.ToLower(uri), "/")
		for _, part := range pathParts {
			if strings.Contains(part, fileContext) {
				score += 25
				break
			}
		}
		
		// File extension match
		ext := strings.ToLower(filepath.Ext(fileName))
		if strings.Contains(ext, fileContext) {
			score += 10
		}

		if score > 0 {
			filtered = append(filtered, symbol)
		}
	}

	return filtered
}

// generateSymbolResponse creates the appropriate response based on results
func generateSymbolResponse(bridge interfaces.BridgeInterface, symbols []SymbolMatch, query, fileContext, detailLevel string) (*mcp.CallToolResult, error) {
	if len(symbols) == 0 {
		message := fmt.Sprintf("No symbols found matching '%s'", query)
		if fileContext != "" {
			message += fmt.Sprintf(" in files containing '%s'", fileContext)
		}
		return mcp.NewToolResultText(message), nil
	}

	// Single match - return detailed information
	if len(symbols) == 1 {
		return generateDetailedSymbolInfo(bridge, symbols[0], detailLevel)
	}

	// Multiple matches - check detail level
	if detailLevel == "full" {
		return generateDetailedMultipleSymbols(bridge, symbols, query, fileContext, detailLevel)
	}

	// Multiple matches - return summary
	return generateSymbolSummary(symbols, query, fileContext)
}

// generateDetailedSymbolInfo gets comprehensive information for a single symbol
func generateDetailedSymbolInfo(bridge interfaces.BridgeInterface, symbol SymbolMatch, detailLevel string) (*mcp.CallToolResult, error) {
	uri := string(symbol.Location.Uri)
	line := symbol.Location.Range.Start.Line
	character := symbol.Location.Range.Start.Character

	var info strings.Builder
	
	// Basic info
	info.WriteString(fmt.Sprintf("Found: %s (%s)\n", symbol.Name, getSymbolKindName(symbol.Kind)))
	info.WriteString(fmt.Sprintf("File: %s:%d\n", filepath.Base(uri), line+1))
	if symbol.ContainerName != "" {
		info.WriteString(fmt.Sprintf("Container: %s\n", symbol.ContainerName))
	}
	info.WriteString("\n")

	// Get hover information for documentation
	if detailLevel == "full" || detailLevel == "auto" {
		hoverInfo, err := bridge.GetHoverInformation(uri, line, character)
		if err == nil && hoverInfo != nil && hoverInfo.Contents.Value != nil {
			info.WriteString("Documentation:\n")
			info.WriteString(formatHoverContent(hoverInfo.Contents))
			info.WriteString("\n\n")
		}

		// Get references count if requested
		references, err := bridge.FindSymbolReferences("", uri, line, character, false)
		if err == nil {
			info.WriteString(fmt.Sprintf("References: %d usages found\n", len(references)))
		}
	}

	return mcp.NewToolResultText(info.String()), nil
}

// generateDetailedMultipleSymbols creates detailed information for multiple symbol matches
func generateDetailedMultipleSymbols(bridge interfaces.BridgeInterface, symbols []SymbolMatch, query, fileContext, detailLevel string) (*mcp.CallToolResult, error) {
	var info strings.Builder
	
	info.WriteString(fmt.Sprintf("Found %d matches for \"%s\"", len(symbols), query))
	if fileContext != "" {
		info.WriteString(fmt.Sprintf(" in files containing \"%s\"", fileContext))
	}
	info.WriteString(" (detailed view):\n\n")

	for i, symbol := range symbols {
		info.WriteString(fmt.Sprintf("=== %d. %s (%s) ===\n", i+1, symbol.Name, getSymbolKindName(symbol.Kind)))
		
		uri := string(symbol.Location.Uri)
		line := symbol.Location.Range.Start.Line
		character := symbol.Location.Range.Start.Character
		
		info.WriteString(fmt.Sprintf("File: %s:%d\n", filepath.Base(uri), line+1))
		if symbol.ContainerName != "" {
			info.WriteString(fmt.Sprintf("Container: %s\n", symbol.ContainerName))
		}
		
		// Get precise coordinates using semantic tokens
		preciseChar := FindPreciseCharacterPosition(bridge, uri, line, character, symbol.Name)
		
		// Try to get hover information using precise coordinates
		hoverInfo, err := bridge.GetHoverInformation(uri, line, preciseChar)
		if err == nil && hoverInfo != nil && hoverInfo.Contents.Value != nil {
			info.WriteString("Documentation:\n")
			info.WriteString(formatHoverContent(hoverInfo.Contents))
			info.WriteString("\n")
		}

		// Try to get references using precise coordinates
		references, err := bridge.FindSymbolReferences("go", uri, line, preciseChar, false)
		if err == nil && len(references) > 0 {
			info.WriteString(fmt.Sprintf("References: %d usages found\n", len(references)))
		}
		
		if i < len(symbols)-1 {
			info.WriteString("\n")
		}
	}

	return mcp.NewToolResultText(info.String()), nil
}

// generateSymbolSummary creates a summary of multiple symbol matches
func generateSymbolSummary(symbols []SymbolMatch, query, fileContext string) (*mcp.CallToolResult, error) {
	var summary strings.Builder
	
	summary.WriteString(fmt.Sprintf("Found %d matches for \"%s\"", len(symbols), query))
	if fileContext != "" {
		summary.WriteString(fmt.Sprintf(" in files containing \"%s\"", fileContext))
	}
	summary.WriteString(":\n\n")

	for i, symbol := range symbols {
		uri := string(symbol.Location.Uri)
		line := symbol.Location.Range.Start.Line
		
		summary.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, symbol.Name, getSymbolKindName(symbol.Kind)))
		summary.WriteString(fmt.Sprintf("   File: %s:%d\n", filepath.Base(uri), line+1))
		if symbol.ContainerName != "" {
			summary.WriteString(fmt.Sprintf("   Container: %s\n", symbol.ContainerName))
		}
		summary.WriteString("\n")
	}

	summary.WriteString("Use file_context parameter to filter results, or specify detail_level=\"full\" for more information.")

	return mcp.NewToolResultText(summary.String()), nil
}

// getSymbolKindName converts SymbolKind to readable string
func getSymbolKindName(kind protocol.SymbolKind) string {
	switch kind {
	case protocol.SymbolKindFile:
		return "file"
	case protocol.SymbolKindModule:
		return "module"
	case protocol.SymbolKindNamespace:
		return "namespace"
	case protocol.SymbolKindPackage:
		return "package"
	case protocol.SymbolKindClass:
		return "class"
	case protocol.SymbolKindMethod:
		return "method"
	case protocol.SymbolKindProperty:
		return "property"
	case protocol.SymbolKindField:
		return "field"
	case protocol.SymbolKindConstructor:
		return "constructor"
	case protocol.SymbolKindEnum:
		return "enum"
	case protocol.SymbolKindInterface:
		return "interface"
	case protocol.SymbolKindFunction:
		return "function"
	case protocol.SymbolKindVariable:
		return "variable"
	case protocol.SymbolKindConstant:
		return "constant"
	case protocol.SymbolKindString:
		return "string"
	case protocol.SymbolKindNumber:
		return "number"
	case protocol.SymbolKindBoolean:
		return "boolean"
	case protocol.SymbolKindArray:
		return "array"
	case protocol.SymbolKindObject:
		return "object"
	case protocol.SymbolKindKey:
		return "key"
	case protocol.SymbolKindNull:
		return "null"
	case protocol.SymbolKindEnumMember:
		return "enum member"
	case protocol.SymbolKindStruct:
		return "struct"
	case protocol.SymbolKindEvent:
		return "event"
	case protocol.SymbolKindOperator:
		return "operator"
	case protocol.SymbolKindTypeParameter:
		return "type parameter"
	default:
		return "symbol"
	}
}

// RegisterSymbolExploreTool registers the symbol exploration tool
func RegisterSymbolExploreTool(mcpServer ToolServer, bridge interfaces.BridgeInterface) {
	mcpServer.AddTool(SymbolExploreTool(bridge))
}