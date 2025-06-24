#!/usr/bin/env python3
import sys
import json
import re
import subprocess
import logging
import uuid

# Configure logging
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(levelname)s: %(message)s"
)
logger = logging.getLogger(__name__)


class MCPToolRunner:
    def __init__(self):
        """
        Initialize MCP server process
        """
        self.mcp_process = None

    def start_mcp_server(self):
        """
        Start the MCP server
        """
        logger.info("Starting MCP server...")

        # Build command to start MCP server
        cmd = ["go", "run", "/home/rockerboo/code/mcp-lsp-bridge/main.go"]

        # Start the server process
        self.mcp_process = subprocess.Popen(
            cmd,
            stdout=subprocess.PIPE,
            stdin=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )

        logger.info("MCP server started")

    def stop_mcp_server(self):
        """
        Stop the MCP server process
        """
        if self.mcp_process:
            logger.info("Stopping MCP server...")
            self.mcp_process.terminate()
            try:
                self.mcp_process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self.mcp_process.kill()

            # Capture and log any output
            stdout, stderr = self.mcp_process.communicate()
            if stdout:
                logger.info(f"Server STDOUT: {stdout}")
            if stderr:
                logger.error(f"Server STDERR: {stderr}")

            self.mcp_process = None

    def parse_mcp_command(self, command_str):
        """
        Parse MCP command string into components.
        This is a more robust parser designed to handle various formats.
        Example input:
        ●? lsp:signature_help (MCP)(uri="file:///path/to/project", line="10", character='5')
        lsp:hover(uri="file:///path/to/file.go", line=10, character=5)
        """
        command_str = command_str.strip()
        
        # Regex to capture the tool name and the raw parameter string.
        # It allows optional '● ', 'lsp:', optional ' (MCP)', and flexible spacing.
        # Group 1: tool_name (e.g., signature_help)
        # Group 2: raw_params_string (e.g., uri="...", line=..., character=...)
        match = re.match(
            r'^(?:●\s*)?lsp:(\w+)\s*(?:\(MCP\))?\s*\((.*)\)$',
            command_str
        )
        
        if not match:
            logger.error(f"Failed to match command structure: '{command_str}'")
            raise ValueError(f"Invalid MCP command format: {command_str}")
        
        tool_name = match.group(1)
        params_str = match.group(2)

        logger.info(f"params string: {params_str}")
        
        params = {}
        if params_str: # Only try to parse if there's a parameter string
            # This regex aims to capture key="value" or key='value' pairs.
            # It might be too strict or not correctly handling escaped quotes within values if they existed.
            # Let's try a slightly different approach for capturing quoted strings.
            
            # We need to match:
            # 1. A key (word characters)
            # 2. Optional whitespace, a colon, optional whitespace
            # 3. A value enclosed in either double or single quotes.
            
            # Regex to find key=value pairs, supporting both double and single quotes for values.
            # It captures:
            # Group 1: key (e.g., uri, line)
            # Group 2: value inside double quotes (e.g., "file://...")
            # Group 3: value inside single quotes (e.g., '5')
            
            # Adjusted regex: Try to be more forgiving with spacing around quotes and keys.
            # This regex specifically tries to capture double-quoted strings OR single-quoted strings.
            param_regex = r'(\w+)\s*=\s*(?:"([^"]*)"|\'([^\']*)\')'
            param_matches = re.findall(param_regex, params_str)

            if not param_matches and params_str.strip():
                 logger.warning(f"Could not parse any parameters from: '{params_str}' for command: {command_str}")
            
            for key, double_quoted_value, single_quoted_value in param_matches:
                # Prefer the value from the matched quotes. If one group is empty, use the other.
                params[key] = double_quoted_value if double_quoted_value else single_quoted_value

        logger.info(f"Parsed command: Tool='{tool_name}', Params={params}") # Log parsed params
        return tool_name, params
    def run_mcp_tool(self, tool_name, params):
        """
        Run MCP tool by sending JSON-RPC request via stdio
        """
        # Generate a unique request ID
        request_id = str(uuid.uuid4())

        # Construct the JSON-RPC request
        request = {
            "jsonrpc": "2.0",
            "method": "tools/call",
            "params": {"name": tool_name, "arguments": params},
            "id": request_id,
        }

        # Serialize the request
        request_json = json.dumps(request)

        logger.info(f"Running MCP Tool: {tool_name}")
        logger.info(f"Parameters: {params}")
        logger.info(f"Request: {request_json}")

        try:
            # Communicate with the running MCP server process
            if not self.mcp_process:
                logger.error("MCP server process is not running")
                return False

            # Write request to stdin
            self.mcp_process.stdin.write(request_json + "\n")
            self.mcp_process.stdin.flush()

            # Read response from stdout
            response_data = self.mcp_process.stdout.readline().strip()

            logger.info(f"Raw Response: {response_data}")

            # Validate response is a non-empty JSON
            if not response_data:
                logger.error("Received empty response")
                return False

            # Parse response
            try:
                response = json.loads(response_data)
            except json.JSONDecodeError as e:
                logger.error(f"Failed to parse response JSON: {e}")
                logger.error(f"Response data: {response_data}")
                return False

            # Check for success or error
            if "error" in response:
                logger.error(f"MCP Tool Error: {response['error']}")
                return False

            # Print result if exists
            if "result" in response:
                print(json.dumps(response["result"], indent=2))

            return True

        except Exception as e:
            logger.error(f"Error running MCP tool: {e}")
            return False

    def __enter__(self):
        self.start_mcp_server()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.stop_mcp_server()


def main():
    if len(sys.argv) < 2:
        print(
            'Usage: python test_mcp_tools.py "lsp:tool_name (MCP)(param1=value1, param2=value2)"'
        )
        sys.exit(1)

    command_str = sys.argv[1]

    try:
        with MCPToolRunner() as runner:
            tool_name, params = runner.parse_mcp_command(command_str)
            success = runner.run_mcp_tool(tool_name, params)
            sys.exit(0 if success else 1)
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
