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
			mcp.WithNumber("limit", mcp.Description("Maximum number of detailed results to show (default: 3)"), mcp.Min(1)),
			mcp.WithNumber("offset", mcp.Description("Number of results to skip for detailed view (default: 0)"), mcp.Min(0)),
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
			limit := request.GetInt("limit", 3)
			offset := request.GetInt("offset", 0)

			logger.Info(fmt.Sprintf("Symbol Explore: query=%s, file_context=%s, detail_level=%s, limit=%d, offset=%d", query, fileContext, detailLevel, limit, offset))

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
			return generateSymbolResponse(bridge, filteredSymbols, query, fileContext, detailLevel, limit, offset)
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
func generateSymbolResponse(bridge interfaces.BridgeInterface, symbols []SymbolMatch, query, fileContext, detailLevel string, limit, offset int) (*mcp.CallToolResult, error) {
	if len(symbols) == 0 {
		message := fmt.Sprintf("No symbols found matching '%s'", query)
		if fileContext != "" {
			message += fmt.Sprintf(" in files containing '%s'", fileContext)
		}
		return mcp.NewToolResultText(message), nil
	}

	// Few matches - show details directly
	if len(symbols) <= 3 {
		return generateDetailedMultipleSymbols(bridge, symbols, query, fileContext)
	}

	// Multiple matches - use table of contents approach
	return generateTableOfContentsResponse(bridge, symbols, query, fileContext, detailLevel, limit, offset)
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

	// Try to get enhanced range information from document symbols
	startLine, startChar, endLine, endChar, err := getEnhancedSymbolRange(bridge, symbol)
	if err != nil {
		// Fall back to original range
		startLine = symbol.Location.Range.Start.Line
		startChar = symbol.Location.Range.Start.Character
		endLine = symbol.Location.Range.End.Line
		endChar = symbol.Location.Range.End.Character
	}

	info.WriteString(fmt.Sprintf("Range: line=%d, character=%d to line=%d, character=%d\n",
		startLine, startChar, endLine, endChar))
	info.WriteString("\n")

	// Get hover information for documentation
	if detailLevel == "full" || detailLevel == "auto" {
		hoverInfo, err := bridge.GetHoverInformation(uri, line, character)
		if err == nil && hoverInfo != nil && hoverInfo.Contents.Value != nil {
			info.WriteString("Documentation:\n")
			info.WriteString(formatHoverContent(hoverInfo.Contents))
			info.WriteString("\n\n")
		}

		// Get symbol content from file using enhanced range details
		content, err := getRangeContent(bridge, uri, startLine, startChar, endLine, endChar, false)
		if err == nil && content != "" {
			info.WriteString("Implementation:\n")
			info.WriteString("```\n")
			info.WriteString(content)
			info.WriteString("\n```\n\n")
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
func generateDetailedMultipleSymbols(bridge interfaces.BridgeInterface, symbols []SymbolMatch, query, fileContext string) (*mcp.CallToolResult, error) {
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

		// Try to get enhanced range information from document symbols
		startLine, startChar, endLine, endChar, err := getEnhancedSymbolRange(bridge, symbol)
		if err != nil {
			// Fall back to original range
			startLine = symbol.Location.Range.Start.Line
			startChar = symbol.Location.Range.Start.Character
			endLine = symbol.Location.Range.End.Line
			endChar = symbol.Location.Range.End.Character
		}

		info.WriteString(fmt.Sprintf("File: %s:%d\n", filepath.Base(uri), line+1))
		info.WriteString(fmt.Sprintf("Range: line=%d, character=%d to line=%d, character=%d\n",
			startLine, startChar, endLine, endChar))
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

		// Get symbol content using enhanced range (non-strict → clamps out-of-bounds chars)
		content, err := getRangeContent(bridge, uri, startLine, startChar, endLine, endChar, false)
		if err == nil && content != "" {
			info.WriteString("Implementation:\n")
			info.WriteString("```\n")
			info.WriteString(content)
			info.WriteString("\n```\n")
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
// func generateSymbolSummary(symbols []SymbolMatch, query, fileContext string) (*mcp.CallToolResult, error) {
// 	var summary strings.Builder
//
// 	summary.WriteString(fmt.Sprintf("Found %d matches for \"%s\"", len(symbols), query))
// 	if fileContext != "" {
// 		summary.WriteString(fmt.Sprintf(" in files containing \"%s\"", fileContext))
// 	}
// 	summary.WriteString(":\n\n")
//
// 	for i, symbol := range symbols {
// 		uri := string(symbol.Location.Uri)
// 		line := symbol.Location.Range.Start.Line
//
// 		summary.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, symbol.Name, getSymbolKindName(symbol.Kind)))
// 		summary.WriteString(fmt.Sprintf("   File: %s:%d\n", filepath.Base(uri), line+1))
// 		if symbol.ContainerName != "" {
// 			summary.WriteString(fmt.Sprintf("   Container: %s\n", symbol.ContainerName))
// 		}
// 		summary.WriteString("\n")
// 	}
//
// 	summary.WriteString("Use file_context parameter to filter results, or specify detail_level=\"full\" for more information.")
//
// 	return mcp.NewToolResultText(summary.String()), nil
// }
//
// generateTableOfContentsResponse creates a table of contents with sample detailed entries
func generateTableOfContentsResponse(bridge interfaces.BridgeInterface, symbols []SymbolMatch, query, fileContext, detailLevel string, limit, offset int) (*mcp.CallToolResult, error) {
	var result strings.Builder

	// Header with total count
	result.WriteString(fmt.Sprintf("Found %d matches for \"%s\"", len(symbols), query))
	if fileContext != "" {
		result.WriteString(fmt.Sprintf(" in files containing \"%s\"", fileContext))
	}
	result.WriteString("\n\n")

	// Only show full table of contents on first page (offset=0)
	if offset == 0 {
		result.WriteString("TABLE OF CONTENTS:\n")
		for i, symbol := range symbols {
			uri := string(symbol.Location.Uri)
			line := symbol.Location.Range.Start.Line

			result.WriteString(fmt.Sprintf("%d. %s (%s) - %s:%d",
				i+1, symbol.Name, getSymbolKindName(symbol.Kind),
				filepath.Base(uri), line+1))
			if symbol.ContainerName != "" {
				result.WriteString(fmt.Sprintf(" [%s]", symbol.ContainerName))
			}
			result.WriteString("\n")
		}
		result.WriteString("\n")
	}

	// Determine which symbols to show in detail
	detailStart := offset
	detailEnd := min(offset + limit, len(symbols))

	if detailStart >= len(symbols) {
		result.WriteString(fmt.Sprintf("\nOffset %d is beyond available results (total: %d)", offset, len(symbols)))
		return mcp.NewToolResultText(result.String()), nil
	}

	// Show detailed information for selected range
	if detailLevel != "basic" && limit > 0 {
		if offset == 0 {
			result.WriteString(fmt.Sprintf("\nDETAILED VIEW (showing %d-%d of %d):\n",
				detailStart+1, detailEnd, len(symbols)))
		} else {
			result.WriteString(fmt.Sprintf("DETAILED VIEW (showing %d-%d of %d matches for \"%s\"):\n",
				detailStart+1, detailEnd, len(symbols), query))
		}

		for i := detailStart; i < detailEnd; i++ {
			result.WriteString(fmt.Sprintf("\n=== %d. %s (%s) ===\n",
				i+1, symbols[i].Name, getSymbolKindName(symbols[i].Kind)))

			// Get detailed info for this symbol
			detailResult, err := generateDetailedSymbolInfo(bridge, symbols[i], detailLevel)
			if err != nil {
				result.WriteString(fmt.Sprintf("Error getting details: %v\n", err))
				continue
			}

			// Extract just the details part (skip the "Found:" header since we have our own)
			if len(detailResult.Content) > 0 {
				if textContent, ok := detailResult.Content[0].(mcp.TextContent); ok {
					detailText := textContent.Text
					lines := strings.Split(detailText, "\n")
					if len(lines) > 1 {
						// Skip the first line ("Found: ...") and join the rest
						detailText = strings.Join(lines[1:], "\n")
					}
					result.WriteString(detailText)
				}
			}
		}

		// Pagination info
		result.WriteString("\n")
		if detailEnd < len(symbols) {
			remaining := len(symbols) - detailEnd
			result.WriteString(fmt.Sprintf("... %d more matches available. Use offset=%d to see more.\n",
				remaining, detailEnd))
		}
		if offset > 0 {
			result.WriteString("Use offset=0 to see from the beginning.\n")
		}
	}

	return mcp.NewToolResultText(result.String()), nil
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

// getEnhancedSymbolRange tries to get better range information using semantic tokens
func getEnhancedSymbolRange(bridge interfaces.BridgeInterface, symbol SymbolMatch) (uint32, uint32, uint32, uint32, error) {
	uri := string(symbol.Location.Uri)
	symbolLine := symbol.Location.Range.Start.Line

	// First try semantic tokens approach - more precise and language-agnostic
	enhanced, err := getSemanticTokenRange(bridge, uri, symbol)
	if err == nil {
		logger.Info(fmt.Sprintf("Semantic tokens SUCCESS: found range %d:%d to %d:%d for %s",
			enhanced.Start.Line, enhanced.Start.Character, enhanced.End.Line, enhanced.End.Character, symbol.Name))
		return enhanced.Start.Line, enhanced.Start.Character, enhanced.End.Line, enhanced.End.Character, nil
	}

	logger.Debug(fmt.Sprintf("Semantic tokens failed for %s: %v, falling back to document symbols", symbol.Name, err))

	// Fallback: Try document symbols approach
	docSymbols, err := bridge.GetDocumentSymbols(uri)
	if err != nil {
		logger.Debug(fmt.Sprintf("Document symbols also failed for %s: %v", uri, err))
		// Fall back to original range
		return symbol.Location.Range.Start.Line, symbol.Location.Range.Start.Character,
			symbol.Location.Range.End.Line, symbol.Location.Range.End.Character, nil
	}

	// Search for the matching symbol in document symbols using location-based matching
	for _, docSymbol := range docSymbols {
		docLine := docSymbol.Range.Start.Line

		// Match by location and kind - language agnostic
		if docSymbol.Kind == symbol.Kind && symbolLine >= docLine && symbolLine <= docLine+5 {
			return docSymbol.Range.Start.Line, docSymbol.Range.Start.Character,
				docSymbol.Range.End.Line, docSymbol.Range.End.Character, nil
		}

		// Also check children symbols recursively
		if enhanced := searchChildSymbols(docSymbol.Children, symbol); enhanced != nil {
			return enhanced.Range.Start.Line, enhanced.Range.Start.Character,
				enhanced.Range.End.Line, enhanced.Range.End.Character, nil
		}
	}

	// Fall back to original range if no match found
	return symbol.Location.Range.Start.Line, symbol.Location.Range.Start.Character,
		symbol.Location.Range.End.Line, symbol.Location.Range.End.Character, nil
}

// getSemanticTokenRange uses semantic tokens to find the full range of a symbol
func getSemanticTokenRange(bridge interfaces.BridgeInterface, uri string, symbol SymbolMatch) (*protocol.Range, error) {
	symbolLine := symbol.Location.Range.Start.Line
	symbolChar := symbol.Location.Range.Start.Character

	// Get semantic tokens for a range around the symbol
	// Look a bit before and after to catch function/method boundaries
	startLine := symbolLine
	if symbolLine >= 5 {
		startLine = symbolLine - 5
	}
	endLine := symbolLine + 50 // Look ahead for the end of the method/function

	// Get semantic tokens for methods, functions, and related types
	tokenTypes := []string{"method", "function", "class", "struct", "interface", "variable", "parameter"}
	logger.Debug(fmt.Sprintf("Getting semantic tokens for %s from %d to %d", symbol.Name, startLine, endLine))
	tokens, err := bridge.SemanticTokens(uri, tokenTypes, startLine, 0, endLine, 1000)
	if err != nil {
		logger.Debug(fmt.Sprintf("Semantic tokens failed: %v", err))
		return nil, fmt.Errorf("failed to get semantic tokens: %w", err)
	}
	logger.Debug(fmt.Sprintf("Got %d semantic tokens", len(tokens)))

	// Find the token that matches our symbol
	var matchedToken *types.TokenPosition
	bestScore := -1

	for _, token := range tokens {
		// Look for tokens that could represent our symbol
		if token.Range.Start.Line == symbolLine {
			// Calculate a score based on how well this token matches
			score := 0

			// Prefer exact name matches
			if strings.Contains(token.Text, symbol.Name) {
				score += 100
			}

			// Prefer tokens close to our expected position
			charDiff := int(symbolChar) - int(token.Range.Start.Character)
			if charDiff < 0 {
				charDiff = -charDiff
			}
			if charDiff < 10 {
				score += 50 - charDiff
			}

			// Prefer method/function tokens for method/function symbols
			if (symbol.Kind == protocol.SymbolKindMethod || symbol.Kind == protocol.SymbolKindFunction) &&
				(token.TokenType == "method" || token.TokenType == "function") {
				score += 75
			}

			if score > bestScore {
				bestScore = score
				matchedToken = &token
			}
		}
	}

	if matchedToken == nil {
		return nil, fmt.Errorf("no matching semantic token found for %s", symbol.Name)
	}

	// For functions and methods, try to find the full body range
	if symbol.Kind == protocol.SymbolKindMethod || symbol.Kind == protocol.SymbolKindFunction {
		// Look for the closing brace by expanding the search
		extendedTokens, err := bridge.SemanticTokens(uri, []string{"delimiter", "punctuation"},
			matchedToken.Range.Start.Line, 0, matchedToken.Range.Start.Line+100, 1000)
		if err == nil {
			// Find the opening and closing braces
			var endRange *protocol.Range
			braceDepth := 0
			foundStart := false

			for _, token := range extendedTokens {
				if token.Range.Start.Line >= matchedToken.Range.Start.Line {
					if strings.Contains(token.Text, "{") {
						braceDepth++
						foundStart = true
					}
					if strings.Contains(token.Text, "}") {
						braceDepth--
						if foundStart && braceDepth == 0 {
							endRange = &protocol.Range{
								Start: matchedToken.Range.Start,
								End:   protocol.Position{Line: token.Range.End.Line, Character: token.Range.End.Character},
							}
							break
						}
					}
				}
			}

			if endRange != nil {
				return endRange, nil
			}
		}
	}

	// Fall back to the token's own range
	return &matchedToken.Range, nil
}

// symbolNamesMatch checks if symbol names match, handling different formats
// func symbolNamesMatch(docSymbolName, workspaceSymbolName string) bool {
// 	// Direct match
// 	if docSymbolName == workspaceSymbolName {
// 		return true
// 	}
//
// 	// Handle Go method receiver syntax: "(*Type).Method" vs "Type.Method"
// 	if strings.HasPrefix(docSymbolName, "(*") && strings.HasSuffix(docSymbolName, ")") {
// 		// Extract "Type.Method" from "(*Type).Method"
// 		simplified := strings.TrimPrefix(docSymbolName, "(*")
// 		simplified = strings.TrimSuffix(simplified, ")")
// 		if simplified == workspaceSymbolName {
// 			return true
// 		}
// 	}
//
// 	// Handle case where workspace symbol might have receiver syntax
// 	if strings.HasPrefix(workspaceSymbolName, "(*") && strings.HasSuffix(workspaceSymbolName, ")") {
// 		simplified := strings.TrimPrefix(workspaceSymbolName, "(*")
// 		simplified = strings.TrimSuffix(simplified, ")")
// 		if simplified == docSymbolName {
// 			return true
// 		}
// 	}
//
// 	return false
// }
//
// searchChildSymbols recursively searches for a matching symbol in children
func searchChildSymbols(children []protocol.DocumentSymbol, target SymbolMatch) *protocol.DocumentSymbol {
	for _, child := range children {
		// Match symbols based on location and kind, with flexible name matching
		targetLine := target.Location.Range.Start.Line
		childLine := child.Range.Start.Line

		// Check if this is the same symbol by location and kind
		if child.Kind == target.Kind && targetLine >= childLine && targetLine <= childLine+5 {
			return &child
		}

		// Recurse into grandchildren
		if found := searchChildSymbols(child.Children, target); found != nil {
			return found
		}
	}
	return nil
}

// RegisterSymbolExploreTool registers the symbol exploration tool
func RegisterSymbolExploreTool(mcpServer ToolServer, bridge interfaces.BridgeInterface) {
	mcpServer.AddTool(SymbolExploreTool(bridge))
}
