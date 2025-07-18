package mcpserver

import (
	"rockerboo/mcp-lsp-bridge/interfaces"
	"rockerboo/mcp-lsp-bridge/mcpserver/tools"
)

// Registers all MCP tools with the server
func RegisterAllTools(mcpServer tools.ToolServer, bridge interfaces.BridgeInterface) {
	// Core analysis tools

	// New unified symbol exploration tool
	tools.RegisterSymbolExploreTool(mcpServer, bridge)

	// Disabling lesser used tools
	// tools.RegisterAnalyzeCodeTool(mcpServer, bridge)
	tools.RegisterProjectAnalysisTool(mcpServer, bridge)

	// Disabling lesser used tools
	// tools.RegisterInferLanguageTool(mcpServer, bridge)
	// tools.RegisterProjectLanguageDetectionTool(mcpServer, bridge)

	// LSP connection management
	// Disabling lesser used tools
	tools.RegisterLSPConnectTool(mcpServer, bridge)
	tools.RegisterLSPDisconnectTool(mcpServer, bridge)

	// Code intelligence tools
	tools.RegisterHoverTool(mcpServer, bridge)
	tools.RegisterSignatureHelpTool(mcpServer, bridge)
	// tools.RegisterDiagnosticsTool(mcpServer, bridge)
	tools.RegisterSemanticTokensTool(mcpServer, bridge)

	// Code improvement tools
	tools.RegisterCodeActionsTool(mcpServer, bridge)
	tools.RegisterFormatDocumentTool(mcpServer, bridge)
	tools.RegisterRangeTools(mcpServer, bridge)

	// Advanced navigation tools
	tools.RegisterRenameTool(mcpServer, bridge)
	tools.RegisterImplementationTool(mcpServer, bridge)

	// Call hierarchy tool
	tools.RegisterCallHierarchyTool(mcpServer, bridge)

	// Workspace analysis
	tools.RegisterWorkspaceDiagnosticsTool(mcpServer, bridge)

	// Diagnostic tools
	tools.RegisterMCPLSPBridgeDiagnosticsTool(mcpServer, bridge)
}
