#!/usr/bin/env python3
import sys
import json
import os
import argparse
import logging
from pathlib import Path

# Add the parent directory to sys.path to import other modules
sys.path.insert(0, str(Path(__file__).parent))
from test_mcp_tools import MCPToolRunner
from test_lsp_init import LSPInitTester

# Configure logging
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(levelname)s: %(message)s"
)
logger = logging.getLogger(__name__)


class LSPCommandTester:
    def __init__(self, config_path, workspace_path=None):
        self.config_path = config_path
        self.workspace_path = workspace_path or "/tmp/test_workspace"
        self.init_tester = LSPInitTester(config_path, workspace_path)

    def parse_command_args(self, args_list):
        """Parse command arguments into a dictionary"""
        params = {}
        for arg in args_list:
            if '=' not in arg:
                logger.warning(f"Skipping invalid argument format: {arg}")
                continue

            key, value = arg.split('=', 1)
            key = key.strip()
            value = value.strip().strip('"\'')

            # Try to convert numeric values
            if value.isdigit():
                params[key] = int(value)
            elif value.lower() in ['true', 'false']:
                params[key] = value.lower() == 'true'
            else:
                params[key] = value

        return params

    def test_hover_command(self, runner, file_uri, line, character):
        """Test hover command at specified position"""
        logger.info(f"Testing hover at {file_uri}:{line}:{character}")

        try:
            tool_name, params = runner.parse_mcp_command(
                f'hover (MCP)(uri="{file_uri}", line={line}, character={character})'
            )
            success = runner.run_mcp_tool(tool_name, params)
            return success
        except Exception as e:
            logger.error(f"Hover command failed: {e}")
            return False

    def test_project_analysis_command(self, runner, analysis_type, query):
        """Test project analysis command"""
        logger.info(f"Testing project analysis: {analysis_type} for {query}")

        try:
            tool_name, params = runner.parse_mcp_command(
                f'project_analysis (MCP)(analysis_type="{analysis_type}", query="{query}")'
            )
            success = runner.run_mcp_tool(tool_name, params)
            return success
        except Exception as e:
            logger.error(f"Project analysis command failed: {e}")
            return False

    def test_signature_help_command(self, runner, file_uri, line, character):
        """Test signature help command at specified position"""
        logger.info(f"Testing signature help at {file_uri}:{line}:{character}")

        try:
            tool_name, params = runner.parse_mcp_command(
                f'signature_help (MCP)(uri="{file_uri}", line={line}, character={character})'
            )
            success = runner.run_mcp_tool(tool_name, params)
            return success
        except Exception as e:
            logger.error(f"Signature help command failed: {e}")
            return False

    def test_workspace_diagnostics_command(self, runner, workspace_uri):
        """Test workspace diagnostics command"""
        logger.info(f"Testing workspace diagnostics for {workspace_uri}")

        try:
            tool_name, params = runner.parse_mcp_command(
                f'workspace_diagnostics (MCP)(workspace_uri="{workspace_uri}")'
            )
            success = runner.run_mcp_tool(tool_name, params)
            return success
        except Exception as e:
            logger.error(f"Workspace diagnostics command failed: {e}")
            return False

    def test_custom_command(self, runner, command_name, command_params):
        """Test a custom MCP command with provided parameters"""
        logger.info(f"Testing custom command: {command_name}")

        try:
            success = runner.run_mcp_tool(command_name, command_params)
            return success
        except Exception as e:
            logger.error(f"Custom command failed: {e}")
            return False

    def run_command_test(self, file_path, command, command_args):
        """Run LSP command test after ensuring connection"""
        logger.info("=== Starting LSP Command Test ===")

        # Validate config file exists
        if not os.path.exists(self.config_path):
            logger.error(f"Config file not found: {self.config_path}")
            return False

        # Create test workspace and file
        test_file = self.init_tester.create_test_workspace(file_path)

        # Detect expected language
        expected_language = self.init_tester.test_language_detection(test_file)
        if not expected_language:
            logger.error("Could not determine expected language for file")
            self.init_tester.cleanup()
            return False

        # Start MCP server with custom config
        config_abs_path = os.path.abspath(self.config_path)
        custom_cmd = ["go", "run", "main.go", "--config", config_abs_path]

        try:
            with MCPToolRunner(custom_cmd=custom_cmd) as runner:
                # First establish LSP connection
                connection_success = self.init_tester.test_lsp_connection(runner, expected_language)

                if not connection_success:
                    logger.error("❌ LSP connection failed, cannot test commands")
                    return False

                # Parse command arguments
                parsed_params = self.parse_command_args(command_args)
                file_uri = f"file://{os.path.abspath(test_file)}"

                # Run the specified command
                success = False

                if command.lower() == "hover":
                    if "uri" not in parsed_params:
                        parsed_params["uri"] = file_uri
                    line = parsed_params.get("line", 1)
                    character = parsed_params.get("character", 5)
                    success = self.test_hover_command(runner, parsed_params["uri"], line, character)

                elif command.lower() == "project_analysis":
                    analysis_type = parsed_params.get("analysis_type", "document_symbols")
                    query = parsed_params.get("query", file_uri)
                    success = self.test_project_analysis_command(runner, analysis_type, query)

                elif command.lower() == "signature_help":
                    if "uri" not in parsed_params:
                        parsed_params["uri"] = file_uri
                    line = parsed_params.get("line", 1)
                    character = parsed_params.get("character", 5)
                    success = self.test_signature_help_command(runner, parsed_params["uri"], line, character)

                elif command.lower() == "workspace_diagnostics":
                    workspace_uri = parsed_params.get("workspace_uri", f"file://{os.path.dirname(os.path.abspath(test_file))}")
                    success = self.test_workspace_diagnostics_command(runner, workspace_uri)

                else:
                    # Try as custom command
                    success = self.test_custom_command(runner, command, parsed_params)

                if success:
                    logger.info(f"🎉 Command '{command}' executed successfully!")
                    return True
                else:
                    logger.error(f"❌ Command '{command}' failed")
                    return False

        except Exception as e:
            logger.error(f"❌ Test runner failed: {e}")
            return False
        finally:
            self.init_tester.cleanup()


def main():
    parser = argparse.ArgumentParser(
        description='Test specific LSP commands on connected language servers',
        epilog='''Examples:

  # Test hover command
  %(prog)s --config test_lsp_config.json --file test.js --cmd hover --args "line=1" "character=5"

  # Test project analysis
  %(prog)s --config test_lsp_config.json --file test.ts --cmd project_analysis --args "analysis_type=document_symbols"

  # Test signature help
  %(prog)s --config test_lsp_config.json --file test.js --cmd signature_help --args "line=2" "character=10"

  # Test workspace diagnostics
  %(prog)s --config test_lsp_config.json --file test.go --cmd workspace_diagnostics
'''
    )

    parser.add_argument('--config', required=True,
                       help='Path to LSP configuration file')
    parser.add_argument('--file', required=True,
                       help='Test file name or path (will be created in workspace)')
    parser.add_argument('--workspace',
                       help='Workspace directory path (default: /tmp/test_workspace)')
    parser.add_argument('--cmd', required=True,
                       help='LSP command to test (hover, project_analysis, signature_help, etc.)')
    parser.add_argument('--args', nargs='*', default=[],
                       help='Command arguments in key=value format')

    args = parser.parse_args()

    try:
        tester = LSPCommandTester(args.config, args.workspace)
        success = tester.run_command_test(args.file, args.cmd, args.args)
        sys.exit(0 if success else 1)

    except Exception as e:
        logger.error(f"Test failed with exception: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
