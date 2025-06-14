package main

import (
	"os"
	"path/filepath"

	"rockerboo/mcp-lsp-bridge/bridge"
	"rockerboo/mcp-lsp-bridge/logger"
	"rockerboo/mcp-lsp-bridge/lsp"
	"rockerboo/mcp-lsp-bridge/mcpserver"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Load LSP configuration
	confPath := "lsp_config.json"
	config, err := lsp.LoadLSPConfig(confPath)
	if err != nil {
		panic("Failed to load LSP config: " + err.Error())
	}

	// Configure logging from config
	logConfig := logger.LoggerConfig{
		LogPath:     config.Global.LogPath,
		LogLevel:    config.Global.LogLevel,
		MaxLogFiles: config.Global.MaxLogFiles,
	}

	// Default to temp file if no path specified
	if logConfig.LogPath == "" {
		logConfig.LogPath = filepath.Join(os.TempDir(), "mcp-lsp-bridge.log")
	}

	if err := logger.InitLogger(logConfig); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Close()

	logger.Info("Starting MCP-LSP Bridge...")

	// Create and initialize the bridge
	bridgeInstance := bridge.NewMCPLSPBridge()

	// Setup MCP server with bridge
	mcpServer := mcpserver.SetupMCPServer(bridgeInstance)

	// Store the server reference in the bridge
	bridgeInstance.SetServer(mcpServer)

	// Start MCP server
	logger.Info("Starting MCP server...")
	if err := server.ServeStdio(mcpServer); err != nil {
		logger.Error("MCP server error: " + err.Error())
	}
}
