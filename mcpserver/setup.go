package mcpserver

import (
	"context"

	"rockerboo/mcp-lsp-bridge/logger"
	"rockerboo/mcp-lsp-bridge/lsp"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// BridgeInterface defines the interface that the bridge must implement
type BridgeInterface interface {
	GetClientForLanguageInterface(language string) (any, error)
	InferLanguage(filePath string) (string, error)
	CloseAllClients()
	GetConfig() *lsp.LSPServerConfig
	DetectProjectLanguages(projectPath string) ([]string, error)
	DetectPrimaryProjectLanguage(projectPath string) (string, error)
}

// SetupMCPServer configures the MCP server with AI-powered tools
func SetupMCPServer(bridge BridgeInterface) *server.MCPServer {
	hooks := &server.Hooks{}

	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		logger.Debug("beforeAny:", method, id, message)
	})
	hooks.AddOnSuccess(func(ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
		logger.Info("onSuccess:", method, id, message, result)
	})
	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		logger.Error("onError:", method, id, message, err)
	})
	hooks.AddBeforeInitialize(func(ctx context.Context, id any, message *mcp.InitializeRequest) {
		logger.Debug("beforeInitialize:", id, message)
	})
	hooks.AddOnRequestInitialization(func(ctx context.Context, id any, message any) error {
		logger.Debug("AddOnRequestInitialization:", id, message)
		// authorization verification and other preprocessing tasks are performed.
		return nil
	})
	hooks.AddAfterInitialize(func(ctx context.Context, id any, message *mcp.InitializeRequest, result *mcp.InitializeResult) {
		logger.Info("afterInitialize:", id, message, result)
	})
	hooks.AddAfterCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest, result *mcp.CallToolResult) {
		logger.Debug("afterCallTool:", id, message, result)
	})
	hooks.AddBeforeCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest) {
		logger.Debug("beforeCallTool:", id, message)
	})

	mcpServer := server.NewMCPServer(
		"lsp-bridge-mcp",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithLogging(),
		server.WithHooks(hooks),
	)

	// Register MCP tools for code analysis
	registerAnalyzeCodeTool(mcpServer, bridge)
	registerInferLanguageTool(mcpServer, bridge)
	registerLSPConnectTool(mcpServer, bridge)
	registerLSPDisconnectTool(mcpServer, bridge)
	registerProjectAnalysisTool(mcpServer, bridge)
	registerProjectLanguageDetectionTool(mcpServer, bridge)

	return mcpServer
}
