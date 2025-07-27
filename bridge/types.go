package bridge

import (
	"rockerboo/mcp-lsp-bridge/types"

	"github.com/mark3labs/mcp-go/server"
)

// MCPLSPBridge combines MCP server capabilities with multiple LSP clients
type MCPLSPBridge struct {
	server             *server.MCPServer
	clients            map[types.LanguageServer]types.LanguageClientInterface
	config             types.LSPServerConfigProvider
	allowedDirectories []string
}
