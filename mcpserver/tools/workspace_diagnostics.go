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

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/myleshyson/lsprotocol-go/protocol"
)

// RegisterWorkspaceDiagnosticsTool registers the workspace_diagnostics tool
func RegisterWorkspaceDiagnosticsTool(mcpServer ToolServer, bridge interfaces.BridgeInterface) {
	tool, handler := WorkspaceDiagnosticsTool(bridge)
	mcpServer.AddTool(tool, handler)
}

func WorkspaceDiagnosticsTool(bridge interfaces.BridgeInterface) (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("workspace_diagnostics",
		mcp.WithDescription("Get comprehensive diagnostics for entire workspace"),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithString("workspace_uri", mcp.Description("URI to the workspace/project root")),
		mcp.WithString("identifier", mcp.Description("Optional identifier for diagnostic session")), // TODO: Add optional when supported
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Parameter parsing
		workspaceUri, err := request.RequireString("workspace_uri")
		if err != nil {
			logger.Error("workspace_diagnostics: workspace_uri parsing failed", err)
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Strip file:// prefix if present
		if after, ok := strings.CutPrefix(workspaceUri, "file://"); ok {
			workspaceUri = after
			logger.Info("workspace_diagnostics: stripped file:// prefix",
				"Processed URI: "+workspaceUri)
		}

		// Optional identifier
		identifier := "mcp-lsp-bridge-workspace-diagnostics"
		if id, err := request.RequireString("identifier"); err == nil {
			identifier = id
		}

		// Detect project languages
		languages, err := bridge.DetectProjectLanguages(workspaceUri)
		if err != nil {
			logger.Error("workspace_diagnostics: language detection failed", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to detect project languages: %v", err)), nil
		}

		if len(languages) == 0 {
			return mcp.NewToolResultText("No programming languages detected in project"), nil
		}

		// Convert languages to strings
		languageStrings := collections.ToString(languages)

		// Get language clients
		clients, err := bridge.GetMultiLanguageClients(languageStrings)
		if err != nil || len(clients) == 0 {
			logger.Error("workspace_diagnostics: failed to get language clients", err)
			return mcp.NewToolResultError("No LSP clients available for detected languages"), nil
		}

		// Convert clients to async operations
		ops := collections.TransformMap(clients, func(client types.LanguageClientInterface) func() (*protocol.WorkspaceDiagnosticReport, error) {
			return func() (*protocol.WorkspaceDiagnosticReport, error) {
				return client.WorkspaceDiagnostic(identifier)
			}
		})

		// Execute diagnostics across all clients in parallel
		results, err := async.MapWithKeys(ctx, ops)
		if err != nil {
			logger.Error("workspace_diagnostics: async execution failed", err)
			return mcp.NewToolResultError(fmt.Sprintf("Failed to execute workspace diagnostics: %v", err)), nil
		}

		// Process results and extract core Diagnostic items
		var allDiagnostics []protocol.Diagnostic
		var languageResults []LanguageDiagnosticResult
		var errors []error

		for _, result := range results {
			if result.Error != nil {
				errors = append(errors, fmt.Errorf("language %s: %w", result.Key, result.Error))
				logger.Warn(fmt.Sprintf("Workspace diagnostics failed for %s: %v", result.Key, result.Error))
			} else if result.Value != nil {
				// Extract diagnostics from the workspace report
				diagnostics := extractDiagnosticsFromWorkspaceReport(result.Value)
				allDiagnostics = append(allDiagnostics, diagnostics...)
				languageResults = append(languageResults, LanguageDiagnosticResult{
					Language:    string(result.Key),
					Diagnostics: diagnostics,
				})
			}
		}

		// Log errors for debugging
		for _, err := range errors {
			logger.Error("workspace_diagnostics: language server error", err)
		}

		// Format results for user-friendly output
		formattedResult := formatWorkspaceDiagnosticsByLanguage(languageResults, allDiagnostics, errors)

		return mcp.NewToolResultText(formattedResult), nil
	}
}

// LanguageDiagnosticResult holds diagnostics for a specific language
type LanguageDiagnosticResult struct {
	Language    string
	Diagnostics []protocol.Diagnostic
}

// extractDiagnosticsFromWorkspaceReport extracts core Diagnostic items from a WorkspaceDiagnosticReport
func extractDiagnosticsFromWorkspaceReport(report *protocol.WorkspaceDiagnosticReport) []protocol.Diagnostic {
	var diagnostics []protocol.Diagnostic
	
	for _, item := range report.Items {
		// Handle the union type - it can be either WorkspaceFullDocumentDiagnosticReport or WorkspaceUnchangedDocumentDiagnosticReport
		switch v := item.Value.(type) {
		case protocol.WorkspaceFullDocumentDiagnosticReport:
			// This contains the actual diagnostics
			diagnostics = append(diagnostics, v.Items...)
		case protocol.WorkspaceUnchangedDocumentDiagnosticReport:
			// This indicates no changes since last check - no new diagnostics to add
			continue
		}
	}
	
	return diagnostics
}

// formatWorkspaceDiagnosticsByLanguage formats workspace diagnostics organized by language
func formatWorkspaceDiagnosticsByLanguage(languageResults []LanguageDiagnosticResult, allDiagnostics []protocol.Diagnostic, errors []error) string {
	var result strings.Builder
	
	// Header with summary
	fmt.Fprintf(&result, "=== WORKSPACE DIAGNOSTICS ===\n")
	fmt.Fprintf(&result, "Languages analyzed: %d\n", len(languageResults))
	fmt.Fprintf(&result, "Total diagnostics: %d\n", len(allDiagnostics))
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
	
	// Group diagnostics by severity for summary
	severityCounts := make(map[protocol.DiagnosticSeverity]int)
	for _, diag := range allDiagnostics {
		if diag.Severity != nil {
			severityCounts[*diag.Severity]++
		} else {
			severityCounts[protocol.DiagnosticSeverityError]++ // Default to error if no severity
		}
	}
	
	// Show severity summary
	if len(allDiagnostics) > 0 {
		fmt.Fprintf(&result, "=== SUMMARY BY SEVERITY ===\n")
		if count, exists := severityCounts[protocol.DiagnosticSeverityError]; exists {
			fmt.Fprintf(&result, "🔴 Errors: %d\n", count)
		}
		if count, exists := severityCounts[protocol.DiagnosticSeverityWarning]; exists {
			fmt.Fprintf(&result, "🟡 Warnings: %d\n", count)
		}
		if count, exists := severityCounts[protocol.DiagnosticSeverityInformation]; exists {
			fmt.Fprintf(&result, "🔵 Information: %d\n", count)
		}
		if count, exists := severityCounts[protocol.DiagnosticSeverityHint]; exists {
			fmt.Fprintf(&result, "💡 Hints: %d\n", count)
		}
		fmt.Fprintf(&result, "\n")
	}
	
	// Show results by language
	if len(languageResults) > 0 {
		fmt.Fprintf(&result, "=== RESULTS BY LANGUAGE ===\n")
		for i, langResult := range languageResults {
			fmt.Fprintf(&result, "%d. %s: %d diagnostics\n", i+1, langResult.Language, len(langResult.Diagnostics))
			
			// Show a few sample diagnostics per language
			sampleCount := min(3, len(langResult.Diagnostics))
			for j := 0; j < sampleCount; j++ {
				diag := langResult.Diagnostics[j]
				severity := "Unknown"
				icon := "❓"
				if diag.Severity != nil {
					severity = getDiagnosticSeverityString(diag.Severity)
					icon = getSeverityIcon(severity)
				}
				fmt.Fprintf(&result, "   %s %s: %s\n", icon, severity, diag.Message)
			}
			
			if len(langResult.Diagnostics) > sampleCount {
				fmt.Fprintf(&result, "   ... and %d more\n", len(langResult.Diagnostics)-sampleCount)
			}
			fmt.Fprintf(&result, "\n")
		}
	} else {
		fmt.Fprintf(&result, "✅ No diagnostics found across all languages\n")
	}
	
	return result.String()
}
