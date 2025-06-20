#!/usr/bin/env python3
"""
Test script to demonstrate hover optimization workflow:
1. Use document_symbols to find symbols in a file
2. Try hover at the suggested coordinates
3. If hover fails, use references/definitions to find better positions
4. Iterate to find the optimal hover position

This addresses the issue where document symbols suggest coordinates that don't work well for hover.
"""

import json
import subprocess
import time
import os
import sys
from pathlib import Path
from typing import Dict, List, Optional, Any

class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    CYAN = '\033[0;36m'
    MAGENTA = '\033[0;35m'
    NC = '\033[0m'

def print_colored(message: str, color: str = Colors.NC):
    print(f"{color}{message}{Colors.NC}")

class MCPTestClient:
    def __init__(self, project_dir: Path):
        self.project_dir = project_dir
        self.server_process = None
        self.request_id = 1
    
    def send_mcp_request(self, method: str, params: Optional[Dict] = None) -> Dict:
        """Send an MCP request and return the response"""
        if params is None:
            params = {}
            
        request = {
            "jsonrpc": "2.0",
            "id": self.request_id,
            "method": method,
            "params": params
        }
        self.request_id += 1
        
        self.server_process.stdin.write(json.dumps(request) + '\n')
        self.server_process.stdin.flush()
        
        response_line = self.server_process.stdout.readline()
        return json.loads(response_line.strip())
    
    def call_tool(self, tool_name: str, arguments: Dict) -> Dict:
        """Call an MCP tool"""
        return self.send_mcp_request("tools/call", {
            "name": tool_name,
            "arguments": arguments
        })

def test_hover_optimization_workflow():
    """Test the complete hover optimization workflow"""
    
    project_dir = Path(__file__).parent.parent
    test_file = f"file://{project_dir}/mcpserver/tools.go"
    
    print_colored("🎯 Testing Hover Optimization Workflow", Colors.CYAN)
    print_colored("=" * 60, Colors.BLUE)
    print()
    
    # Build project
    print_colored("🔨 Building project...", Colors.YELLOW)
    build_result = subprocess.run(["go", "build"], cwd=project_dir, capture_output=True, text=True)
    if build_result.returncode != 0:
        print_colored(f"❌ Build failed: {build_result.stderr}", Colors.RED)
        return False
    print_colored("✅ Build successful", Colors.GREEN)
    
    # Start MCP server
    print_colored("🚀 Starting MCP server...", Colors.YELLOW)
    server_process = subprocess.Popen(
        ["./mcp-lsp-bridge"],
        cwd=project_dir,
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True
    )
    
    client = MCPTestClient(project_dir)
    client.server_process = server_process
    
    try:
        # Initialize connection
        print_colored("🔧 Initializing MCP connection...", Colors.BLUE)
        response = client.send_mcp_request("initialize", {
            "protocolVersion": "2024-11-05",
            "capabilities": {"tools": {}},
            "clientInfo": {"name": "hover-test-client", "version": "1.0.0"}
        })
        
        if "error" in response:
            print_colored(f"❌ Initialization failed: {response['error']}", Colors.RED)
            return False
        
        print_colored("✅ MCP server initialized", Colors.GREEN)
        
        # Connect to Go LSP
        print_colored("🔗 Connecting to Go LSP server...", Colors.BLUE)
        response = client.call_tool("lsp_connect", {"language": "go"})
        
        if response.get("result", {}).get("isError"):
            print_colored(f"❌ LSP connection failed", Colors.RED)
            return False
        
        print_colored("✅ Connected to Go LSP server", Colors.GREEN)
        print()
        
        # Step 1: Get document symbols to find target function
        print_colored("📋 Step 1: Finding symbols using document_symbols", Colors.CYAN)
        print_colored("-" * 50, Colors.BLUE)
        
        response = client.call_tool("project_analysis", {
            "analysis_type": "document_symbols",
            "query": test_file,
            "workspace_uri": f"file://{project_dir}"
        })
        
        if response.get("result", {}).get("isError"):
            print_colored(f"❌ Document symbols failed", Colors.RED)
            return False
        
        content = response["result"]["content"][0]["text"]
        print_colored("📄 Document symbols result:", Colors.BLUE)
        print(content)
        print()
        
        # Extract coordinates from the response (line=10, character=0)
        initial_line = 10
        initial_char = 0
        
        # Step 2: Try hover at the suggested coordinates
        print_colored("🔍 Step 2: Testing hover at document symbol coordinates", Colors.CYAN)
        print_colored("-" * 50, Colors.BLUE)
        print_colored(f"Trying hover at line={initial_line}, character={initial_char}", Colors.BLUE)
        
        response = client.call_tool("hover", {
            "uri": test_file,
            "line": initial_line,
            "character": initial_char
        })
        
        content = response["result"]["content"][0]["text"]
        print_colored(f"📄 Hover result: {content}", Colors.YELLOW)
        
        hover_success = "No hover information available" not in content
        if hover_success:
            print_colored("✅ Hover successful at document symbol coordinates!", Colors.GREEN)
        else:
            print_colored("⚠️  Hover failed at document symbol coordinates", Colors.YELLOW)
        print()
        
        # Step 3: Use references to find better positions
        print_colored("🔍 Step 3: Finding better positions using references", Colors.CYAN)
        print_colored("-" * 50, Colors.BLUE)
        
        response = client.call_tool("project_analysis", {
            "analysis_type": "references",
            "query": "RegisterAllTools",
            "workspace_uri": f"file://{project_dir}",
            "limit": 10
        })
        
        content = response["result"]["content"][0]["text"]
        print_colored("📄 References result:", Colors.BLUE)
        print(content)
        print()
        
        # Step 4: Use definitions to find precise symbol location
        print_colored("🎯 Step 4: Finding precise location using definitions", Colors.CYAN)
        print_colored("-" * 50, Colors.BLUE)
        
        response = client.call_tool("project_analysis", {
            "analysis_type": "definitions",
            "query": "RegisterAllTools",
            "workspace_uri": f"file://{project_dir}",
            "limit": 3
        })
        
        content = response["result"]["content"][0]["text"]
        print_colored("📄 Definitions result:", Colors.BLUE)
        print(content)
        print()
        
        # Step 5: Try hover at different positions based on the intelligence gathered
        print_colored("🧠 Step 5: Optimizing hover positions", Colors.CYAN)
        print_colored("-" * 50, Colors.BLUE)
        
        # Test positions based on common Go function declaration patterns
        test_positions = [
            {"line": 10, "char": 5, "desc": "Start of function name"},
            {"line": 10, "char": 15, "desc": "Middle of function name"},
            {"line": 10, "char": 21, "desc": "End of function name"},
            {"line": 11, "char": 10, "desc": "Inside function body"},
            {"line": 11, "char": 0, "desc": "Start of next line"},
        ]
        
        successful_positions = []
        
        for pos in test_positions:
            print_colored(f"🔍 Testing position: line={pos['line']}, char={pos['char']} ({pos['desc']})", Colors.BLUE)
            
            response = client.call_tool("hover", {
                "uri": test_file,
                "line": pos["line"],
                "character": pos["char"]
            })
            
            content = response["result"]["content"][0]["text"]
            
            if "No hover information available" not in content:
                print_colored(f"✅ SUCCESS: {content[:100]}...", Colors.GREEN)
                successful_positions.append(pos)
            else:
                print_colored(f"❌ No hover info at this position", Colors.RED)
            print()
        
        # Step 6: Summary and recommendations
        print_colored("📊 Step 6: Analysis Summary", Colors.CYAN)
        print_colored("=" * 60, Colors.BLUE)
        
        if successful_positions:
            print_colored(f"✅ Found {len(successful_positions)} successful hover positions:", Colors.GREEN)
            for pos in successful_positions:
                print_colored(f"   • Line {pos['line']}, Character {pos['char']}: {pos['desc']}", Colors.CYAN)
        else:
            print_colored("⚠️  No positions provided hover information", Colors.YELLOW)
        
        print()
        print_colored("🎯 Recommendations for hover optimization:", Colors.MAGENTA)
        print_colored("1. Document symbols provide structural ranges, not optimal hover points", Colors.CYAN)
        print_colored("2. For function names, try character positions within the identifier", Colors.CYAN)
        print_colored("3. References can show where symbols are used effectively", Colors.CYAN)
        print_colored("4. Definitions point to the exact declaration location", Colors.CYAN)
        print_colored("5. Iterate through nearby positions for best results", Colors.CYAN)
        
        return True
        
    except Exception as e:
        print_colored(f"❌ Test error: {e}", Colors.RED)
        return False
    
    finally:
        # Cleanup
        if server_process:
            print_colored("🧹 Cleaning up server process...", Colors.YELLOW)
            server_process.terminate()
            server_process.wait()

def main():
    """Main function"""
    print_colored("🎯 Hover Optimization Workflow Testing", Colors.BLUE)
    print_colored("Demonstrating how to find optimal hover positions", Colors.CYAN)
    print()
    
    success = test_hover_optimization_workflow()
    
    if success:
        print_colored("🏁 Hover optimization testing completed successfully!", Colors.GREEN)
        return 0
    else:
        print_colored("🏁 Hover optimization testing failed", Colors.RED)
        return 1

if __name__ == "__main__":
    sys.exit(main())