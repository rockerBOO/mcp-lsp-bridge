# MCP Tools Reference

Complete reference for all MCP tools provided by the LSP bridge.

## Quick Reference

**Analysis**: `project_analysis`, `symbol_explore`, `workspace_diagnostics`
**Navigation**: `hover`, `implementation`, `call_hierarchy`
**Refactoring**: `rename`, `format_document`, `code_actions`
**Utilities**: `detect_project_languages`, `get_range_content`, `infer_language`

## Core Analysis Tools

### `project_analysis`
Multi-purpose code analysis with 9 analysis types for symbols, files, and workspace patterns.

**Common Usage:**
- Find symbols: `analysis_type="workspace_symbols"`, `query="calculateTotal"`
- Analyze files: `analysis_type="file_analysis"`, `query="src/auth.go"`
- Workspace overview: `analysis_type="workspace_analysis"`, `query="entire_project"`

**Key Parameters**: analysis_type (required), query (required), limit (default: 20)
**Output**: Structured analysis results with metadata and suggestions

### `symbol_explore`
Intelligent symbol search with contextual filtering and detailed code information.

**Common Usage:**
- Find symbols: `query="getUserData"`
- Filter by context: `query="validateUser"`, `file_context="auth"`
- Detailed view: `query="connectDB"`, `detail_level="full"`

**Key Parameters**: query (required), file_context (optional), detail_level (auto/basic/full)
**Output**: Symbol matches with documentation, implementation, and references

### `get_range_content`
Extract text content from specific file ranges with precise line/character positioning.

**Common Usage:**
- Extract function: `uri="file://path"`, `start_line=10`, `end_line=25`
- Get code block: Use coordinates from `project_analysis` definitions

**Key Parameters**: uri (required), start_line/end_line (required), strict (default: false)
**Output**: Exact text content from specified range

### `analyze_code`
Analyze code for completion suggestions and insights.

## Language Detection Tools

### `detect_project_languages`
Detect all programming languages used in a project by examining root markers and file extensions.

### `infer_language`
Infer the programming language for a file.

## LSP Connection Management

### `lsp_connect`
Connect to a language server for a specific language.

### `lsp_disconnect`
Disconnect all active language server clients.

## Code Intelligence Tools

### `hover`
Get detailed symbol information including signatures, documentation, and type details.

**Common Usage:**
- Symbol info: `uri="file://path"`, `line=15`, `character=10`
- Type details: Position cursor on variable/function name

**Key Parameters**: uri (required), line/character (required, 0-based)
**Output**: Formatted documentation with code examples and pkg.go.dev links

### `signature_help`
Get function parameter information at call sites. Use when positioned inside function calls (between parentheses) to see parameter details and overloads.

### `semantic_tokens`
Get semantic tokens for a specific range of a file.

### `workspace_diagnostics`
Analyze entire workspace for errors, warnings, and code issues across all languages.

**Common Usage:**
- Full scan: `workspace_uri="file://project/root"`
- Check health: Review error categories and language-specific issues

**Key Parameters**: workspace_uri (required)
**Output**: Categorized diagnostics by language with error explanations and suggestions

## Code Improvement & Refactoring Tools

### `code_actions`
Get intelligent quick fixes, refactoring suggestions, and automated code improvements.

**Common Usage:**
- Fix errors: Position at error location for import fixes, syntax corrections
- Refactor: Position at symbol for extract method, implement interface options

**Key Parameters**: uri (required), line/character (required)
**Output**: Available actions with descriptions and edit previews

### `format_document`
Format documents with language-specific rules. Preview changes before applying.

**Common Usage:**
- Preview: `uri="file://path"`, `apply="false"` (default) - shows changes
- Apply: `uri="file://path"`, `apply="true"` - formats file

**Key Parameters**: uri (required), apply (default: false), tab_size (default: 4)
**Output**: Formatting changes with line-by-line diffs

### `rename`
Rename symbols across entire codebase with cross-file precision. Always preview first.

**Common Usage:**
- Preview: `uri="file://path"`, `line=10`, `character=5`, `new_name="newFunc"`, `apply="false"`
- Apply: Same parameters with `apply="true"`

**Key Parameters**: uri (required), line/character (required), new_name (required), apply (default: false)
**Output**: All affected files with exact change locations

## Advanced Navigation Tools

### `implementation`
Find implementations of a symbol (interfaces, abstract methods).

### `call_hierarchy`
Show call hierarchy (callers and callees) for a symbol.

## Common Workflows

**Code Analysis**: `detect_project_languages` → `project_analysis` → `symbol_explore`
**Navigation**: `hover` → `implementation` → `call_hierarchy`
**Refactoring**: `hover` (get coordinates) → `rename` (preview) → `rename` (apply)
**Code Quality**: `workspace_diagnostics` → `code_actions` → `format_document`

## System Tools

### `mcp_lsp_diagnostics`
Provides diagnostic information about the MCP-LSP bridge, including registered language servers, configuration details, connected servers, and detected project languages.

## Safety Features

For tools that modify code (`format_document`, `rename`), the bridge provides crucial safety mechanisms:

- **Preview Mode**: Shows exactly what changes will be made across all affected files without modifying them
- **Apply Mode**: Once reviewed and approved, applies the changes to your codebase

This dual-mode operation ensures full control and visibility over automated code modifications.

## Tool Categories Summary

**Analysis & Discovery**: `project_analysis`, `symbol_explore`, `get_range_content`, `analyze_code`

**Language Detection**: `detect_project_languages`, `infer_language`

**Connection Management**: `lsp_connect`, `lsp_disconnect`

**Code Intelligence**: `hover`, `signature_help`, `semantic_tokens`, `workspace_diagnostics`

**Safe Refactoring**: `code_actions`, `format_document`, `rename` (with preview/apply modes)

**Navigation**: `implementation`, `call_hierarchy`

**System**: `mcp_lsp_diagnostics`
