#!/usr/bin/env python3
"""
MCP External Testing Script
Tests the MCP-LSP Bridge server by simulating external client interactions
"""

import json
import subprocess
import time
import os
import sys
import signal
import threading
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Any

class Colors:
    """ANSI color codes for terminal output"""
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    NC = '\033[0m'  # No Color

class TestResult:
    """Represents the result of a test"""
    def __init__(self, name: str, description: str):
        self.name = name
        self.description = description
        self.success = False
        self.error = ""
        self.duration = 0.0
        self.response = None
        self.start_time = None
        self.end_time = None

class MCPTestClient:
    """External MCP test client"""
    
    def __init__(self, project_dir: Path):
        self.project_dir = project_dir
        self.scripts_dir = project_dir / "scripts"
        self.build_output = project_dir / "mcp-lsp-bridge"
        self.log_file = self.scripts_dir / "mcp_test.log"
        self.report_file = project_dir / "mcp_test_report.json"
        
        self.server_process = None
        self.request_id = 1
        self.tests_total = 0
        self.tests_passed = 0
        self.tests_failed = 0
        
        # Ensure scripts directory exists
        self.scripts_dir.mkdir(exist_ok=True)
        
    def log_message(self, message: str):
        """Log a message with timestamp"""
        timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        with open(self.log_file, "a") as f:
            f.write(f"{timestamp} - {message}\n")
    
    def print_colored(self, message: str, color: str = Colors.NC):
        """Print a colored message"""
        print(f"{color}{message}{Colors.NC}")
    
    def build_project(self) -> bool:
        """Build the MCP-LSP Bridge project"""
        self.print_colored("🔨 Building MCP-LSP Bridge...", Colors.YELLOW)
        
        try:
            result = subprocess.run(
                ["go", "build", "-o", str(self.build_output), "."],
                cwd=self.project_dir,
                capture_output=True,
                text=True,
                timeout=60
            )
            
            if result.returncode == 0:
                self.print_colored("✅ Build successful", Colors.GREEN)
                self.log_message("Build successful")
                return True
            else:
                self.print_colored("❌ Build failed", Colors.RED)
                self.print_colored(f"   Error: {result.stderr}", Colors.RED)
                self.log_message(f"Build failed: {result.stderr}")
                return False
                
        except subprocess.TimeoutExpired:
            self.print_colored("❌ Build timed out", Colors.RED)
            self.log_message("Build timed out")
            return False
        except Exception as e:
            self.print_colored(f"❌ Build error: {e}", Colors.RED)
            self.log_message(f"Build error: {e}")
            return False
    
    def start_server(self) -> bool:
        """Start the MCP server"""
        self.print_colored("🚀 Starting MCP server...", Colors.YELLOW)
        
        try:
            # Start server process
            self.server_process = subprocess.Popen(
                [str(self.build_output)],
                cwd=self.project_dir,
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )
            
            # Wait a moment for server to start
            time.sleep(2)
            
            # Check if server is still running
            if self.server_process.poll() is None:
                self.print_colored(f"✅ MCP server started (PID: {self.server_process.pid})", Colors.GREEN)
                self.log_message(f"MCP server started with PID: {self.server_process.pid}")
                return True
            else:
                self.print_colored("❌ MCP server failed to start", Colors.RED)
                stderr_output = self.server_process.stderr.read() if self.server_process.stderr else ""
                self.log_message(f"MCP server failed to start: {stderr_output}")
                return False
                
        except Exception as e:
            self.print_colored(f"❌ Failed to start server: {e}", Colors.RED)
            self.log_message(f"Failed to start server: {e}")
            return False
    
    def send_request(self, method: str, params: Optional[Dict] = None) -> TestResult:
        """Send a JSON-RPC request to the MCP server"""
        if params is None:
            params = {}
            
        request = {
            "jsonrpc": "2.0",
            "id": self.request_id,
            "method": method,
            "params": params
        }
        self.request_id += 1
        
        test_result = TestResult(method, f"Method: {method}")
        test_result.start_time = time.time()
        
        try:
            if not self.server_process or self.server_process.poll() is not None:
                raise Exception("Server process is not running")
            
            # Send request
            request_json = json.dumps(request) + "\n"
            self.server_process.stdin.write(request_json)
            self.server_process.stdin.flush()
            
            # Read response with timeout using select or polling
            import select
            import sys
            
            # Wait for response with timeout
            ready, _, _ = select.select([self.server_process.stdout], [], [], 10.0)
            
            if not ready:
                raise Exception("Request timed out - no response received")
            
            # Read the response line
            response_line = self.server_process.stdout.readline()
            if not response_line.strip():
                raise Exception("Empty response received")
            
            response = json.loads(response_line.strip())
            test_result.response = response
            
            # Check for JSON-RPC error
            if "error" in response:
                raise Exception(f"Server returned error: {response['error']}")
            
            test_result.success = True
            
        except Exception as e:
            test_result.success = False
            test_result.error = str(e)
        
        test_result.end_time = time.time()
        test_result.duration = test_result.end_time - test_result.start_time
        
        return test_result
    
    def test_tool_call(self, tool_name: str, arguments: Dict) -> TestResult:
        """Test calling a specific MCP tool"""
        params = {
            "name": tool_name,
            "arguments": arguments
        }
        
        result = self.send_request("tools/call", params)
        result.name = f"tool_{tool_name}"
        result.description = f"Tool: {tool_name}"
        
        return result
    
    def run_test(self, test_result: TestResult):
        """Run a single test and update counters"""
        self.tests_total += 1
        
        self.print_colored(f"🔍 Testing: {test_result.description}", Colors.BLUE)
        self.log_message(f"Starting test: {test_result.name}")
        
        if test_result.success:
            self.print_colored(f"✅ Success: {test_result.description} ({test_result.duration:.3f}s)", Colors.GREEN)
            self.log_message(f"Test passed: {test_result.name} in {test_result.duration:.3f}s")
            self.tests_passed += 1
        else:
            self.print_colored(f"❌ Failed: {test_result.description} ({test_result.duration:.3f}s)", Colors.RED)
            self.print_colored(f"   Error: {test_result.error}", Colors.RED)
            self.log_message(f"Test failed: {test_result.name} in {test_result.duration:.3f}s")
            self.log_message(f"Error: {test_result.error}")
            self.tests_failed += 1
    
    def run_all_tests(self) -> List[TestResult]:
        """Run all MCP tests"""
        self.print_colored("🧪 Running Comprehensive MCP Tool Tests...", Colors.YELLOW)
        print()
        
        results = []
        
        # Test 1: Initialize
        result = self.send_request("initialize", {
            "protocolVersion": "2024-11-05",
            "capabilities": {"tools": {}},
            "clientInfo": {"name": "test-client", "version": "1.0.0"}
        })
        result.description = "Initialize MCP Connection"
        self.run_test(result)
        results.append(result)
        
        # Send initialized notification
        if result.success:
            notification = {
                "jsonrpc": "2.0",
                "method": "notifications/initialized"
            }
            try:
                notification_json = json.dumps(notification) + "\n"
                self.server_process.stdin.write(notification_json)
                self.server_process.stdin.flush()
            except Exception as e:
                self.log_message(f"Failed to send initialized notification: {e}")
        
        # Test 2: List Tools
        result = self.send_request("tools/list")
        result.description = "List Available Tools"
        self.run_test(result)
        results.append(result)
        
        # Core Analysis Tools
        print()
        self.print_colored("🔍 Testing Core Analysis Tools", Colors.BLUE)
        
        # Test 3: Infer Language Tool
        result = self.test_tool_call("infer_language", {"file_path": "/test/example.go"})
        self.run_test(result)
        results.append(result)
        
        # Test 4: Detect Project Languages Tool
        result = self.test_tool_call("detect_project_languages", {
            "project_path": str(self.project_dir),
            "mode": "all"
        })
        self.run_test(result)
        results.append(result)
        
        # Test 5: Project Analysis Tool - Workspace Symbols
        result = self.test_tool_call("project_analysis", {
            "workspace_uri": f"file://{self.project_dir}",
            "analysis_type": "workspace_symbols",
            "query": "RegisterAllTools",
            "limit": 5
        })
        self.run_test(result)
        results.append(result)
        
        # Test 6: Project Analysis Tool - Document Symbols
        result = self.test_tool_call("project_analysis", {
            "workspace_uri": f"file://{self.project_dir}",
            "analysis_type": "document_symbols",
            "query": f"file://{self.project_dir}/mcpserver/tools.go",
            "limit": 10
        })
        self.run_test(result)
        results.append(result)
        
        # LSP Connection Management
        print()
        self.print_colored("🔗 Testing LSP Connection Management", Colors.BLUE)
        
        # Test 7: LSP Connect Tool
        result = self.test_tool_call("lsp_connect", {"language": "go"})
        self.run_test(result)
        results.append(result)
        
        # Code Intelligence Tools (require active LSP connection)
        print()
        self.print_colored("🧠 Testing Code Intelligence Tools", Colors.BLUE)
        
        # Test 8: Analyze Code Tool
        result = self.test_tool_call("analyze_code", {
            "uri": f"file://{self.project_dir}/mcpserver/tools.go",
            "line": 11,
            "character": 5
        })
        self.run_test(result)
        results.append(result)
        
        # Test 9: Hover Tool
        result = self.test_tool_call("hover", {
            "uri": f"file://{self.project_dir}/mcpserver/tools.go",
            "line": 11,
            "character": 5
        })
        self.run_test(result)
        results.append(result)
        
        # Test 10: Signature Help Tool (NEW - recently fixed)
        result = self.test_tool_call("signature_help", {
            "uri": f"file://{self.project_dir}/mcpserver/tools.go",
            "line": 13,
            "character": 40  # Position after function call opening parenthesis
        })
        self.run_test(result)
        results.append(result)
        
        # Test 11: Diagnostics Tool
        result = self.test_tool_call("diagnostics", {
            "uri": f"file://{self.project_dir}/mcpserver/tools.go"
        })
        self.run_test(result)
        results.append(result)
        
        # Code Improvement Tools
        print()
        self.print_colored("🛠️ Testing Code Improvement Tools", Colors.BLUE)
        
        # Test 12: Code Actions Tool
        result = self.test_tool_call("code_actions", {
            "uri": f"file://{self.project_dir}/mcpserver/tools.go",
            "line": 11,
            "character": 0,
            "end_line": 11,
            "end_character": 50
        })
        self.run_test(result)
        results.append(result)
        
        # Test 13: Format Document Tool
        result = self.test_tool_call("format_document", {
            "uri": f"file://{self.project_dir}/mcpserver/tools.go",
            "tab_size": 4
        })
        self.run_test(result)
        results.append(result)
        
        # Advanced Navigation Tools
        print()
        self.print_colored("🧭 Testing Advanced Navigation Tools", Colors.BLUE)
        
        # Test 14: Implementation Tool (NEW - recently fixed)
        result = self.test_tool_call("implementation", {
            "uri": f"file://{self.project_dir}/mcpserver/tools.go",
            "line": 11,
            "character": 19  # Position on function name "RegisterAllTools"
        })
        self.run_test(result)
        results.append(result)
        
        # Test 15: Rename Tool
        result = self.test_tool_call("rename", {
            "uri": f"file://{self.project_dir}/mcpserver/tools.go",
            "line": 11,
            "character": 19,
            "new_name": "RegisterAllMCPTools",
            "preview": True  # Preview mode to avoid actual changes
        })
        self.run_test(result)
        results.append(result)
        
        # Test 16: Call Hierarchy Tool
        result = self.test_tool_call("call_hierarchy", {
            "uri": f"file://{self.project_dir}/mcpserver/tools.go",
            "line": 11,
            "character": 19,
            "direction": "both"
        })
        self.run_test(result)
        results.append(result)
        
        # Workspace Analysis Tools
        print()
        self.print_colored("🌐 Testing Workspace Analysis Tools", Colors.BLUE)
        
        # Test 17: Workspace Diagnostics Tool
        result = self.test_tool_call("workspace_diagnostics", {
            "workspace_uri": f"file://{self.project_dir}",
            "identifier": "test-session"
        })
        self.run_test(result)
        results.append(result)
        
        # Test 18: Project Analysis - References
        result = self.test_tool_call("project_analysis", {
            "workspace_uri": f"file://{self.project_dir}",
            "analysis_type": "references",
            "query": "RegisterAllTools",
            "limit": 5
        })
        self.run_test(result)
        results.append(result)
        
        # Test 19: Project Analysis - Definitions
        result = self.test_tool_call("project_analysis", {
            "workspace_uri": f"file://{self.project_dir}",
            "analysis_type": "definitions",
            "query": "RegisterAllTools",
            "limit": 3
        })
        self.run_test(result)
        results.append(result)
        
        # Test 20: Project Analysis - Text Search
        result = self.test_tool_call("project_analysis", {
            "workspace_uri": f"file://{self.project_dir}",
            "analysis_type": "text_search",
            "query": "MCP tools",
            "limit": 5
        })
        self.run_test(result)
        results.append(result)
        
        # Special Integration Tests
        print()
        self.print_colored("🎯 Testing Hover Optimization Workflow", Colors.BLUE)
        
        # Test 21: Document symbols + hover optimization workflow
        # This demonstrates the exact workflow mentioned by the user
        result = self.test_tool_call("project_analysis", {
            "analysis_type": "document_symbols",
            "query": "mcpserver/tools.go",
            "workspace_uri": f"file://{self.project_dir}"
        })
        result.description = "Document Symbols for Hover Optimization"
        self.run_test(result)
        results.append(result)
        
        # Test 22: Hover at document symbol coordinates (expected to fail)
        result = self.test_tool_call("hover", {
            "uri": f"file://{self.project_dir}/mcpserver/tools.go",
            "line": 10,
            "character": 0
        })
        result.description = "Hover at Document Symbol Coordinates (line=10, char=0)"
        self.run_test(result)
        results.append(result)
        
        # Test 23: Hover at optimized position (expected to succeed)
        result = self.test_tool_call("hover", {
            "uri": f"file://{self.project_dir}/mcpserver/tools.go",
            "line": 10,
            "character": 5  # Inside the function name
        })
        result.description = "Hover at Optimized Position (line=10, char=5)"
        self.run_test(result)
        results.append(result)
        
        # Cleanup - LSP Disconnect
        print()
        self.print_colored("🧹 Testing Cleanup", Colors.BLUE)
        
        # Test 24: LSP Disconnect Tool
        result = self.test_tool_call("lsp_disconnect", {})
        self.run_test(result)
        results.append(result)
        
        return results
    
    def generate_report(self, results: List[TestResult]):
        """Generate test report"""
        print()
        print("=" * 60)
        self.print_colored("📋 MCP EXTERNAL TEST REPORT", Colors.BLUE)
        print("=" * 60)
        
        print("📊 Summary:")
        print(f"   Total Tests: {self.tests_total}")
        self.print_colored(f"   Passed: {self.tests_passed}", Colors.GREEN)
        self.print_colored(f"   Failed: {self.tests_failed}", Colors.RED)
        
        if self.tests_failed == 0:
            self.print_colored("🎉 All tests passed!", Colors.GREEN)
        else:
            self.print_colored(f"⚠️  {self.tests_failed} test(s) failed", Colors.RED)
        
        print()
        print(f"📄 Log file: {self.log_file}")
        
        # Generate JSON report
        report_data = {
            "timestamp": datetime.now().isoformat(),
            "summary": {
                "total": self.tests_total,
                "passed": self.tests_passed,
                "failed": self.tests_failed
            },
            "tests": [
                {
                    "name": result.name,
                    "description": result.description,
                    "success": result.success,
                    "error": result.error,
                    "duration": result.duration,
                    "response": result.response
                }
                for result in results
            ],
            "log_file": str(self.log_file),
            "project_dir": str(self.project_dir)
        }
        
        with open(self.report_file, "w") as f:
            json.dump(report_data, f, indent=2)
        
        print(f"📄 Report file: {self.report_file}")
        print()
    
    def cleanup(self):
        """Clean up resources"""
        if self.server_process:
            self.print_colored("🧹 Cleaning up MCP server process...", Colors.YELLOW)
            try:
                self.server_process.terminate()
                self.server_process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self.server_process.kill()
                self.server_process.wait()
            except Exception as e:
                self.log_message(f"Error during cleanup: {e}")
        
        # Remove build output
        if self.build_output.exists():
            try:
                self.build_output.unlink()
            except Exception as e:
                self.log_message(f"Error removing build output: {e}")

def main():
    """Main function"""
    # Determine project directory
    script_dir = Path(__file__).parent
    project_dir = script_dir.parent
    
    print(f"{Colors.BLUE}🚀 Starting MCP External Testing{Colors.NC}")
    print("=" * 50)
    print(f"Project Directory: {project_dir}")
    print(f"Scripts Directory: {script_dir}")
    print()
    
    client = MCPTestClient(project_dir)
    
    # Set up signal handler for cleanup
    def signal_handler(signum, frame):
        print(f"\n{Colors.YELLOW}Received signal {signum}, cleaning up...{Colors.NC}")
        client.cleanup()
        sys.exit(1)
    
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    try:
        # Initialize log file
        with open(client.log_file, "w") as f:
            f.write(f"{datetime.now().strftime('%Y-%m-%d %H:%M:%S')} - Starting MCP external testing suite\n")
        
        # Build project
        if not client.build_project():
            print(f"{Colors.RED}❌ Build failed, exiting{Colors.NC}")
            return 1
        
        # Start MCP server
        if not client.start_server():
            print(f"{Colors.RED}❌ Failed to start MCP server, exiting{Colors.NC}")
            return 1
        
        # Run tests
        results = client.run_all_tests()
        
        # Generate report
        client.generate_report(results)
        
        # Determine exit code
        if client.tests_failed == 0:
            print(f"{Colors.GREEN}🏁 Testing completed successfully{Colors.NC}")
            return 0
        else:
            print(f"{Colors.RED}🏁 Testing completed with failures{Colors.NC}")
            return 1
            
    except Exception as e:
        print(f"{Colors.RED}❌ Unexpected error: {e}{Colors.NC}")
        client.log_message(f"Unexpected error: {e}")
        return 1
    
    finally:
        client.cleanup()

if __name__ == "__main__":
    sys.exit(main())