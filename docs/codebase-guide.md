# Codebase Guide

Comprehensive overview of the MCP-LSP Bridge codebase structure and functionality for LLM agents.

## 🏗️ Project Architecture

**MCP-LSP Bridge** is a Go application that bridges Model Context Protocol (MCP) and Language Server Protocol (LSP), exposing LSP capabilities as MCP tools for AI agents.

**Core Flow:** MCP Client → MCP Server (this project) → LSP Clients → Language Servers

## 📁 Directory Structure

### 🚀 Entry Point & Core

- **`main.go`** - Application entry point, command-line parsing, initialization
- **`go.mod`/`go.sum`** - Go module dependencies
- **`Makefile`** - Build scripts and common tasks

### 🌉 Bridge Layer

- **`bridge/`** - Core LSP-MCP bridge interface
  - `bridge.go` - Main bridge implementation connecting MCP to LSP
  - `bridge_test.go` - Bridge functionality tests
  - `types.go` - Bridge-specific type definitions

- **`interfaces/`** - Go interfaces for loose coupling
  - `bridge.go` - Bridge interface definitions

### 🔌 LSP Client Layer

- **`lsp/`** - Language Server Protocol client implementation
  - `client.go` - Main LSP client implementation
  - `config.go` - LSP server configuration management
  - `handler.go` - LSP message handlers
  - `methods.go` - LSP method implementations
  - `semantic_tokens.go` - Semantic token analysis
  - `types.go` - LSP-specific types
  - `mocks/` - Mock implementations for testing

### 🛠️ MCP Server Layer

- **`mcpserver/`** - Model Context Protocol server implementation
  - `setup.go` - MCP server configuration and initialization
  - `tools.go` - Tool registry and management
  - `tools/` - Individual MCP tool implementations (16 tools)
    - `project_analysis.go` - Multi-purpose code analysis tool
    - `symbol_explore.go` - Intelligent symbol search
    - `hover.go` - Symbol information retrieval
    - `code_actions.go` - Quick fixes and refactoring suggestions
    - `format_document.go` - Code formatting with preview/apply modes
    - `rename.go` - Safe symbol renaming across codebase
    - `workspace_diagnostics.go` - Project-wide error detection
    - `lsp_connect.go`/`lsp_disconnect.go` - LSP connection management
    - And 8 more specialized tools...

### 🔍 Analysis Engine

- **`analysis/`** - Advanced code analysis capabilities
  - `engine.go` - Core analysis engine with 9 analysis types
  - `cache.go` - Analysis result caching for performance
  - `performance.go` - Performance monitoring and metrics
  - `types.go` - Analysis-specific type definitions
  - `errors.go` - Analysis error handling

### 🧰 Utility Modules

- **`async/`** - Asynchronous operation management
  - `async.go` - Async operations for concurrent LSP calls

- **`collections/`** - Data structure utilities
  - `collections.go` - Generic collection helpers

- **`utils/`** - General utility functions
  - `uri.go` - URI handling and validation
  - `flatten.go` - Data flattening utilities

- **`types/`** - Shared type definitions
  - `lsp.go` - LSP-related types
  - `lsp_client_metrics.go` - Performance metrics types

- **`logger/`** - Centralized logging
  - `logger.go` - Structured logging implementation

- **`security/`** - Security utilities
  - `path_validation.go` - Path validation and sanitization

- **`directories/`** - Directory and path management
  - `directories.go` - Platform-specific directory handling

### 🧪 Testing & Quality

- **`mocks/`** - Mock implementations for testing
  - `bridge.go`, `lsp_client.go`, etc. - Interface mocks

- **Test Coverage:**
  - `*_test.go` files throughout codebase
  - `coverage*` files - Test coverage reports
  - `unit_coverage.out` - Unit test coverage data

### 📚 Documentation & Scripts

- **`docs/`** - User and developer documentation
  - `tools/` - Tool-specific documentation
  - `configuration.md` - Setup guide
  - `analysis-overview.md` - Quick start guide

- **`scripts/`** - Development and testing scripts
  - `test_mcp_tools.py` - End-to-end MCP tool testing
  - `performance_test.py` - Performance benchmarking
  - `memory_test.py` - Memory usage analysis

- **`notes/`** - Development notes and progress tracking
  - Daily development logs (YYYY-MM-DD.md format)
  - Comprehensive test reports
  - Implementation progress tracking

### 📖 Configuration & Examples

- **`lsp_config.example.json`** - Example LSP server configuration
- **`Dockerfile`** - Container configuration

### 🧬 LSP Protocol Documentation

- **`lsp_parsed/`** - Parsed LSP specification documentation
  - Organized LSP protocol reference
  - Implementation guidance

## 🔄 Key Data Flow

1. **MCP Request** → `mcpserver/tools/*.go`
2. **Tool Processing** → `bridge/bridge.go`
3. **LSP Communication** → `lsp/client.go`
4. **Analysis (if needed)** → `analysis/engine.go`
5. **Response** → Back through the chain

## 🎯 Key Components for LLM Agents

### When to look where:

- **Adding new MCP tools** → `mcpserver/tools/`
- **LSP protocol issues** → `lsp/`
- **Analysis features** → `analysis/`
- **Configuration problems** → `lsp/config.go`, `main.go`
- **Testing** → `*_test.go` files, `scripts/`
- **Documentation** → `docs/`
- **Architecture understanding** → `bridge/`, `interfaces/`

### Critical Files for Understanding:

1. **`main.go`** - How everything starts
2. **`bridge/bridge.go`** - Core bridge logic
3. **`mcpserver/tools/project_analysis.go`** - Most complex tool
4. **`analysis/engine.go`** - Analysis capabilities
5. **`lsp/client.go`** - LSP communication

## 🏷️ Naming Conventions

- **Packages**: Lowercase, descriptive (`bridge`, `analysis`, `mcpserver`)
- **Files**: Snake_case for multi-word concepts (`project_analysis.go`)
- **Functions**: PascalCase for public, camelCase for private
- **Constants**: ALL_CAPS with underscores
- **Interfaces**: Usually end with `Interface` (`BridgeInterface`)

## 🔧 Build & Development

- **Build**: `go build` or `make`
- **Test**: `go test ./...` or `make test`
- **Lint**: `make lint`
- **Coverage**: Generated in `coverage*` files

This codebase follows Go best practices with clean separation of concerns, comprehensive testing, and extensive documentation.
