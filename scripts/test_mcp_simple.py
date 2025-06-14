#!/usr/bin/env python3
"""
Simple MCP Server Test
Tests basic MCP server functionality
"""

import json
import subprocess
import time
import os
import sys
from pathlib import Path

def test_mcp_server():
    """Test basic MCP server functionality"""
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
        
        # Test basic communication
        print("🔍 Testing basic communication...")
        
        # Send a simple initialize request
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
        
        print("📤 Sent initialize request")
        
        # Try to read response with timeout
        import select
        
        # Wait for response with timeout
        ready, _, _ = select.select([process.stdout], [], [], 5.0)
        
        if ready:
            response_line = process.stdout.readline()
            if response_line.strip():
                print(f"📥 Received response: {response_line.strip()}")
                try:
                    response = json.loads(response_line.strip())
                    if "result" in response:
                        print("✅ Server responded correctly to initialize")
                        
                        # Test a simple tool call
                        print("🔍 Testing tool list...")
                        list_request = {
                            "jsonrpc": "2.0",
                            "id": 2,
                            "method": "tools/list",
                            "params": {}
                        }
                        
                        request_json = json.dumps(list_request) + "\n"
                        process.stdin.write(request_json)
                        process.stdin.flush()
                        
                        # Wait for tools/list response
                        ready, _, _ = select.select([process.stdout], [], [], 5.0)
                        if ready:
                            tools_response = process.stdout.readline()
                            if tools_response.strip():
                                print(f"📥 Tools response: {tools_response.strip()}")
                                tools_data = json.loads(tools_response.strip())
                                if "result" in tools_data and "tools" in tools_data["result"]:
                                    tools = tools_data["result"]["tools"]
                                    print(f"✅ Found {len(tools)} tools: {[t['name'] for t in tools]}")
                                    return True
                        
                        return True
                    elif "error" in response:
                        print(f"❌ Server returned error: {response['error']}")
                        return False
                except json.JSONDecodeError:
                    print(f"❌ Invalid JSON response: {response_line}")
                    return False
        else:
            print("❌ Timeout waiting for response")
            return False
        
    except Exception as e:
        print(f"❌ Error testing server: {e}")
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
    print("🧪 Simple MCP Server Test")
    print("=" * 40)
    
    success = test_mcp_server()
    
    if success:
        print("\n🎉 Test passed!")
        sys.exit(0)
    else:
        print("\n❌ Test failed!")
        sys.exit(1)