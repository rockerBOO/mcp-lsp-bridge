#!/bin/bash

# MCP External Testing Script
# Tests the MCP-LSP Bridge server by simulating external client interactions

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
TEST_TIMEOUT=30
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPTS_DIR="$PROJECT_DIR/scripts"
BUILD_OUTPUT="$PROJECT_DIR/mcp-lsp-bridge"
LOG_FILE="$SCRIPTS_DIR/mcp_test.log"
REPORT_FILE="$PROJECT_DIR/mcp_test_report.json"

# Counters
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0

echo -e "${BLUE}🚀 Starting MCP External Testing${NC}"
echo "=================================================="
echo "Project Directory: $PROJECT_DIR"
echo "Scripts Directory: $SCRIPTS_DIR"
echo "Build Output: $BUILD_OUTPUT"
echo "Log File: $LOG_FILE"
echo ""

# Cleanup function
cleanup() {
    if [ ! -z "$MCP_PID" ]; then
        echo -e "${YELLOW}🧹 Cleaning up MCP server process...${NC}"
        kill $MCP_PID 2>/dev/null || true
        wait $MCP_PID 2>/dev/null || true
    fi
    if [ -f "$BUILD_OUTPUT" ]; then
        rm -f "$BUILD_OUTPUT"
    fi
}

# Set up trap for cleanup
trap cleanup EXIT

# Function to log messages
log_message() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# Function to run a test
run_test() {
    local test_name="$1"
    local test_description="$2"
    local mcp_request="$3"
    
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    echo -e "${BLUE}🔍 Testing: $test_description${NC}"
    log_message "Starting test: $test_name"
    
    # Send the MCP request via stdin and capture response
    local start_time=$(date +%s.%N)
    local response
    local exit_code=0
    
    # Use timeout to prevent hanging
    if response=$(timeout $TEST_TIMEOUT bash -c "echo '$mcp_request' | nc -q 1 localhost 12345 2>/dev/null" 2>&1); then
        local end_time=$(date +%s.%N)
        local duration=$(echo "$end_time - $start_time" | bc -l)
        
        echo -e "${GREEN}✅ Success: $test_description (${duration}s)${NC}"
        log_message "Test passed: $test_name in ${duration}s"
        log_message "Response: $response"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        local end_time=$(date +%s.%N)
        local duration=$(echo "$end_time - $start_time" | bc -l)
        
        echo -e "${RED}❌ Failed: $test_description (${duration}s)${NC}"
        echo -e "${RED}   Error: $response${NC}"
        log_message "Test failed: $test_name in ${duration}s"
        log_message "Error: $response"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Function to build the project
build_project() {
    echo -e "${YELLOW}🔨 Building MCP-LSP Bridge...${NC}"
    cd "$PROJECT_DIR"
    
    if go build -o "$BUILD_OUTPUT" .; then
        echo -e "${GREEN}✅ Build successful${NC}"
        return 0
    else
        echo -e "${RED}❌ Build failed${NC}"
        exit 1
    fi
}

# Function to start MCP server
start_mcp_server() {
    echo -e "${YELLOW}🚀 Starting MCP server...${NC}"
    cd "$PROJECT_DIR"
    
    # Start the server in background
    "$BUILD_OUTPUT" > "$LOG_FILE" 2>&1 &
    MCP_PID=$!
    
    # Wait a moment for server to start
    sleep 2
    
    # Check if server is still running
    if kill -0 $MCP_PID 2>/dev/null; then
        echo -e "${GREEN}✅ MCP server started (PID: $MCP_PID)${NC}"
        log_message "MCP server started with PID: $MCP_PID"
        return 0
    else
        echo -e "${RED}❌ MCP server failed to start${NC}"
        log_message "MCP server failed to start"
        return 1
    fi
}

# Function to test server readiness
test_server_readiness() {
    echo -e "${YELLOW}🔗 Testing server readiness...${NC}"
    
    # Try to connect to server
    local retries=10
    local retry_count=0
    
    while [ $retry_count -lt $retries ]; do
        if nc -z localhost 12345 2>/dev/null; then
            echo -e "${GREEN}✅ Server is ready${NC}"
            return 0
        fi
        
        retry_count=$((retry_count + 1))
        echo "   Attempt $retry_count/$retries..."
        sleep 1
    done
    
    echo -e "${RED}❌ Server not ready after $retries attempts${NC}"
    return 1
}

# Function to run MCP JSON-RPC tests
run_mcp_tests() {
    echo -e "${YELLOW}🧪 Running MCP Tool Tests...${NC}"
    echo ""
    
    # Test 1: Initialize
    local init_request='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{"tools":{}},"clientInfo":{"name":"test-client","version":"1.0.0"}}}'
    run_test "initialize" "Initialize MCP Connection" "$init_request"
    
    # Test 2: List Tools
    local list_tools_request='{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
    run_test "list_tools" "List Available Tools" "$list_tools_request"
    
    # Test 3: Infer Language Tool
    local infer_language_request='{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"infer_language","arguments":{"file_path":"/test/example.go"}}}'
    run_test "infer_language" "Infer Language Tool" "$infer_language_request"
    
    # Test 4: LSP Connect Tool
    local lsp_connect_request='{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"lsp_connect","arguments":{"language":"go"}}}'
    run_test "lsp_connect" "LSP Connect Tool" "$lsp_connect_request"
    
    # Test 5: Analyze Code Tool
    local analyze_code_request='{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"analyze_code","arguments":{"uri":"file:///test/example.go","line":10,"character":5}}}'
    run_test "analyze_code" "Analyze Code Tool" "$analyze_code_request"
    
    # Test 6: LSP Disconnect Tool
    local lsp_disconnect_request='{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"lsp_disconnect","arguments":{}}}'
    run_test "lsp_disconnect" "LSP Disconnect Tool" "$lsp_disconnect_request"
}

# Function to generate test report
generate_report() {
    echo ""
    echo "============================================================"
    echo -e "${BLUE}📋 MCP EXTERNAL TEST REPORT${NC}"
    echo "============================================================"
    
    echo -e "📊 Summary:"
    echo -e "   Total Tests: $TESTS_TOTAL"
    echo -e "   Passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "   Failed: ${RED}$TESTS_FAILED${NC}"
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}🎉 All tests passed!${NC}"
    else
        echo -e "${RED}⚠️  $TESTS_FAILED test(s) failed${NC}"
    fi
    
    echo ""
    echo -e "📄 Log file: $LOG_FILE"
    
    # Generate JSON report
    cat > "$REPORT_FILE" << EOF
{
  "timestamp": "$(date -Iseconds)",
  "summary": {
    "total": $TESTS_TOTAL,
    "passed": $TESTS_PASSED,
    "failed": $TESTS_FAILED
  },
  "log_file": "$LOG_FILE",
  "project_dir": "$PROJECT_DIR"
}
EOF
    
    echo -e "📄 Report file: $REPORT_FILE"
    echo ""
}

# Function to check dependencies
check_dependencies() {
    echo -e "${YELLOW}🔍 Checking dependencies...${NC}"
    
    # Check for required commands
    local deps=("go" "nc" "bc" "timeout")
    local missing_deps=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing_deps+=("$dep")
        fi
    done
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        echo -e "${RED}❌ Missing dependencies: ${missing_deps[*]}${NC}"
        echo "Please install the missing dependencies and try again."
        exit 1
    fi
    
    echo -e "${GREEN}✅ All dependencies found${NC}"
}

# Main execution
main() {
    echo -e "${BLUE}Starting MCP External Testing Suite${NC}"
    echo ""
    
    # Initialize log file
    > "$LOG_FILE"
    log_message "Starting MCP external testing suite"
    
    # Check dependencies
    check_dependencies
    
    # Build project
    build_project
    
    # Start MCP server
    if ! start_mcp_server; then
        echo -e "${RED}❌ Failed to start MCP server${NC}"
        exit 1
    fi
    
    # Test server readiness
    if ! test_server_readiness; then
        echo -e "${RED}❌ Server readiness test failed${NC}"
        exit 1
    fi
    
    # Run MCP tests
    run_mcp_tests
    
    # Generate report
    generate_report
    
    # Determine exit code
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}🏁 Testing completed successfully${NC}"
        exit 0
    else
        echo -e "${RED}🏁 Testing completed with failures${NC}"
        exit 1
    fi
}

# Run main function
main "$@"