# MCP External Testing Scripts

This directory contains scripts for testing the MCP-LSP Bridge server externally, simulating how Claude Code would interact with the MCP server.

## Available Scripts

### 1. Python Test Script (Recommended)
- **File**: `test_mcp_external.py`
- **Description**: Comprehensive Python-based external MCP testing
- **Usage**: `python3 test_mcp_external.py` or `make test-mcp-external`

**Features:**
- JSON-RPC communication with MCP server
- Comprehensive test coverage for all MCP tools
- Detailed error handling and reporting
- Colored terminal output
- JSON report generation
- Automatic cleanup

### 2. Shell Script
- **File**: `test_mcp_external.sh`
- **Description**: Bash-based external MCP testing
- **Usage**: `./test_mcp_external.sh` or `make test-mcp-external-shell`

**Features:**
- Cross-platform shell scripting
- Basic MCP server interaction
- Simple reporting
- Dependency checking

### 3. Simple MCP Test
- **File**: `test_mcp_simple.py`
- **Description**: Basic MCP connectivity and tool discovery test
- **Usage**: `python3 test_mcp_simple.py` or `make test-mcp-simple`

**Features:**
- Quick connectivity verification
- Tool discovery testing
- Basic protocol validation

### 4. MCP Tools Test
- **File**: `test_mcp_tools.py`
- **Description**: Individual MCP tools functionality test
- **Usage**: `python3 test_mcp_tools.py` or `make test-mcp-tools`

**Features:**
- Tests each MCP tool individually
- Validates tool responses
- Comprehensive functionality verification

### 5. Go Test Client
- **File**: `test_mcp_external.go`
- **Description**: Go-based MCP test client (advanced)
- **Usage**: `go run test_mcp_external.go`

**Features:**
- Native Go implementation
- Full MCP protocol support
- Advanced test scenarios

## Test Coverage

The external tests cover the following MCP tools:

1. **Initialize**: MCP connection initialization
2. **List Tools**: Enumerate available MCP tools
3. **Infer Language**: Test language detection for files
4. **LSP Connect**: Test language server connection
5. **Analyze Code**: Test code analysis functionality
6. **LSP Disconnect**: Test cleanup and disconnection

## Requirements

### Python Script
- Python 3.6+
- Standard library only (no external dependencies)

### Shell Script
- Bash 4.0+
- `nc` (netcat)
- `bc` (basic calculator)
- `timeout` command

### Go Script
- Go 1.19+
- MCP-Go library

## Usage Examples

```bash
# Run simple connectivity test (recommended for quick verification)
make test-mcp-simple

# Run individual tools test
make test-mcp-tools

# Run comprehensive external tests
make test-mcp-external

# Run shell-based tests
make test-mcp-external-shell

# Run tests directly
cd scripts
python3 test_mcp_simple.py
python3 test_mcp_tools.py
python3 test_mcp_external.py
```

## Output Files

- **Log File**: `scripts/mcp_test.log` - Detailed test execution log
- **Report File**: `mcp_test_report.json` - JSON test report with results

## Test Results

Each test produces:
- **Success/Failure status**
- **Execution duration**
- **Error messages** (if any)
- **Server responses** (in JSON report)

## Integration with CI/CD

The tests are designed to be CI/CD friendly:
- Return appropriate exit codes (0 = success, 1 = failure)
- Generate machine-readable JSON reports
- Provide colored output for human readability
- Include timeout handling to prevent hanging

## Troubleshooting

### Common Issues

1. **Build Failures**
   - Ensure Go is installed and in PATH
   - Run `go mod tidy` in project root

2. **Server Start Failures**
   - Check if required LSP servers are installed
   - Review log file for detailed error messages

3. **Connection Issues**
   - Ensure no other process is using the same port
   - Check firewall settings

4. **Timeout Issues**
   - Increase timeout values in scripts
   - Check system performance

### Debug Mode

For debugging, check the log file:
```bash
tail -f scripts/mcp_test.log
```

## Contributing

When adding new tests:
1. Follow the existing test pattern
2. Add appropriate error handling
3. Update this README
4. Ensure cross-platform compatibility