#!/usr/bin/env python3
"""
MCP Tools Test
Tests individual MCP tools functionality
"""

import json
import subprocess
import time
import os
import sys
import select
from pathlib import Path

def test_mcp_tools():
    """Test individual MCP tools"""
    project_dir = Path(__file__).parent.parent
    build_output = project_dir / "mcp-lsp-bridge"
    
    print("🔨 Building MCP-LSP Bridge...")
    # Build the project
    result = subprocess.run(
        ["go", "build", "-o", str(build_output), "."],
        cwd=project_dir,
        capture_output=True,
        text=True
    )
    
    if result.returncode != 0:
        print(f"❌ Build failed: {result.stderr}")
        return False
    
    print("✅ Build successful")
    
    print("🚀 Starting MCP server...")
    # Start the server
    try:
        process = subprocess.Popen(
            [str(build_output)],
            cwd=project_dir,
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        
        # Wait a moment for server to start
        time.sleep(1)
        
        # Check if server is running
        if process.poll() is not None:
            stderr_output = process.stderr.read()
            print(f"❌ Server failed to start: {stderr_output}")
            return False
        
        print(f"✅ Server started (PID: {process.pid})")
        
        # Initialize
        print("🔗 Initializing connection...")
        init_request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize",
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {"tools": {}},
                "clientInfo": {"name": "test-client", "version": "1.0.0"}
            }
        }
        
        # Send the request
        request_json = json.dumps(init_request) + "\n"
        process.stdin.write(request_json)
        process.stdin.flush()
        
        # Read response
        ready, _, _ = select.select([process.stdout], [], [], 5.0)
        if not ready:
            print("❌ Initialize timeout")
            return False
        
        response_line = process.stdout.readline()
        print("✅ Initialized successfully")
        
        # Test individual tools
        tests_passed = 0
        tests_total = 0
        
        # Test 1: infer_language
        print("\n🔍 Testing infer_language tool...")
        tests_total += 1
        tool_request = {
            "jsonrpc": "2.0",
            "id": 2,
            "method": "tools/call",
            "params": {
                "name": "infer_language",
                "arguments": {
                    "file_path": "/test/example.go"
                }
            }
        }
        
        request_json = json.dumps(tool_request) + "\n"
        process.stdin.write(request_json)
        process.stdin.flush()
        
        ready, _, _ = select.select([process.stdout], [], [], 5.0)
        if ready:
            response_line = process.stdout.readline()
            if response_line.strip():
                response = json.loads(response_line.strip())
                if "result" in response and response["result"]["content"]:
                    content = response["result"]["content"][0]["text"]
                    print(f"✅ infer_language: {content}")
                    tests_passed += 1
                else:
                    print(f"❌ infer_language failed: {response}")
        else:
            print("❌ infer_language timeout")
        
        # Test 2: lsp_connect
        print("\n🔍 Testing lsp_connect tool...")
        tests_total += 1
        tool_request = {
            "jsonrpc": "2.0",
            "id": 3,
            "method": "tools/call",
            "params": {
                "name": "lsp_connect",
                "arguments": {
                    "language": "go"
                }
            }
        }
        
        request_json = json.dumps(tool_request) + "\n"
        process.stdin.write(request_json)
        process.stdin.flush()
        
        ready, _, _ = select.select([process.stdout], [], [], 5.0)
        if ready:
            response_line = process.stdout.readline()
            if response_line.strip():
                response = json.loads(response_line.strip())
                if "result" in response and response["result"]["content"]:
                    content = response["result"]["content"][0]["text"]
                    print(f"✅ lsp_connect: {content}")
                    tests_passed += 1
                else:
                    print(f"❌ lsp_connect failed: {response}")
        else:
            print("❌ lsp_connect timeout")
        
        # Test 3: lsp_disconnect
        print("\n🔍 Testing lsp_disconnect tool...")
        tests_total += 1
        tool_request = {
            "jsonrpc": "2.0",
            "id": 4,
            "method": "tools/call",
            "params": {
                "name": "lsp_disconnect",
                "arguments": {}
            }
        }
        
        request_json = json.dumps(tool_request) + "\n"
        process.stdin.write(request_json)
        process.stdin.flush()
        
        ready, _, _ = select.select([process.stdout], [], [], 5.0)
        if ready:
            response_line = process.stdout.readline()
            if response_line.strip():
                response = json.loads(response_line.strip())
                if "result" in response and response["result"]["content"]:
                    content = response["result"]["content"][0]["text"]
                    print(f"✅ lsp_disconnect: {content}")
                    tests_passed += 1
                else:
                    print(f"❌ lsp_disconnect failed: {response}")
        else:
            print("❌ lsp_disconnect timeout")
        
        print(f"\n📊 Results: {tests_passed}/{tests_total} tools passed")
        
        return tests_passed == tests_total
        
    except Exception as e:
        print(f"❌ Error testing tools: {e}")
        return False
    
    finally:
        # Cleanup
        try:
            if 'process' in locals():
                process.terminate()
                process.wait(timeout=2)
        except:
            if 'process' in locals():
                process.kill()
        
        # Remove build output
        if build_output.exists():
            build_output.unlink()

if __name__ == "__main__":
    print("🧪 MCP Tools Test")
    print("=" * 40)
    
    success = test_mcp_tools()
    
    if success:
        print("\n🎉 All tools tests passed!")
        sys.exit(0)
    else:
        print("\n❌ Some tools tests failed!")
        sys.exit(1)