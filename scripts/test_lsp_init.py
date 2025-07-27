#!/usr/bin/env python3
import sys
import json
import os
import argparse
import tempfile
import subprocess
import logging
from pathlib import Path

# Add the parent directory to sys.path to import test_mcp_tools
sys.path.insert(0, str(Path(__file__).parent))
from test_mcp_tools import MCPToolRunner

# Configure logging
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(levelname)s: %(message)s"
)
logger = logging.getLogger(__name__)


class LSPInitTester:
    def __init__(self, config_path, workspace_path=None):
        self.config_path = config_path
        self.workspace_path = workspace_path or "/tmp/test_workspace"
        self.temp_files = []

    def create_test_workspace(self, file_path):
        """Create a test workspace with the specified file"""
        # Ensure the workspace directory exists
        os.makedirs(self.workspace_path, exist_ok=True)

        # Determine the full file path
        if not os.path.isabs(file_path):
            full_file_path = os.path.join(self.workspace_path, file_path)
        else:
            full_file_path = file_path

        # Create directory for the file if it doesn't exist
        os.makedirs(os.path.dirname(full_file_path), exist_ok=True)

        # Create the file with some basic content
        file_extension = os.path.splitext(full_file_path)[1]
        if file_extension in ['.js', '.jsx', '.mjs', '.cjs']:
            content = '''// Test JavaScript file
function greet(name) {
    console.log(`Hello, ${name}!`);
}

const message = "Testing LSP";
greet(message);
'''
        elif file_extension in ['.ts', '.tsx']:
            content = '''// Test TypeScript file
interface Person {
    name: string;
    age: number;
}

function greet(person: Person): void {
    console.log(`Hello, ${person.name}!`);
}

const testPerson: Person = { name: "Test", age: 25 };
greet(testPerson);
'''
        elif file_extension == '.go':
            content = '''// Test Go file
package main

import "fmt"

func greet(name string) {
    fmt.Printf("Hello, %s!\\n", name)
}

func main() {
    message := "Testing LSP"
    greet(message)
}
'''
        else:
            content = f"// Test file with extension {file_extension}"

        with open(full_file_path, 'w') as f:
            f.write(content)

        self.temp_files.append(full_file_path)
        logger.info(f"Created test file: {full_file_path}")
        return full_file_path

    def test_language_detection(self, file_path):
        """Test if the file extension is properly detected"""
        extension = os.path.splitext(file_path)[1]
        logger.info(f"Testing file with extension: {extension}")

        # Map extensions to expected languages based on our config
        extension_to_language = {
            '.js': 'typescript',
            '.jsx': 'typescript',
            '.mjs': 'typescript',
            '.cjs': 'typescript',
            '.ts': 'typescript',
            '.tsx': 'typescript',
            '.go': 'go'
        }

        expected_language = extension_to_language.get(extension)
        if expected_language:
            logger.info(f"Expected language for {extension}: {expected_language}")
            return expected_language
        else:
            logger.warning(f"No expected language mapping for extension: {extension}")
            return None

    def test_lsp_connection(self, runner, language):
        """Test LSP connection for the given language"""
        logger.info(f"Testing LSP connection for language: {language}")

        try:
            tool_name, params = runner.parse_mcp_command(f"lsp_connect (MCP)(language=\"{language}\")")
            success = runner.run_mcp_tool(tool_name, params)

            if success:
                logger.info(f"✅ Successfully connected to LSP for {language}")
                return True
            else:
                logger.error(f"❌ Failed to connect to LSP for {language}")
                return False

        except Exception as e:
            logger.error(f"❌ Exception during LSP connection test: {e}")
            return False

    def test_file_analysis(self, runner, file_path):
        """Test basic file analysis to verify the connection works"""
        file_uri = f"file://{os.path.abspath(file_path)}"
        logger.info(f"Testing file analysis for: {file_uri}")

        try:
            # Test document symbols
            tool_name, params = runner.parse_mcp_command(
                f'project_analysis (MCP)(analysis_type="document_symbols", query="{file_uri}")'
            )
            success = runner.run_mcp_tool(tool_name, params)

            if success:
                logger.info(f"✅ Successfully analyzed file: {file_path}")
                return True
            else:
                logger.error(f"❌ Failed to analyze file: {file_path}")
                return False

        except Exception as e:
            logger.error(f"❌ Exception during file analysis: {e}")
            return False

    def cleanup(self):
        """Clean up temporary files"""
        for file_path in self.temp_files:
            try:
                if os.path.exists(file_path):
                    os.remove(file_path)
                    logger.info(f"Cleaned up: {file_path}")
            except Exception as e:
                logger.warning(f"Failed to clean up {file_path}: {e}")

    def run_test(self, file_path):
        """Run the complete LSP initialization test"""
        logger.info("=== Starting LSP Initialization Test ===")

        # Validate config file exists
        if not os.path.exists(self.config_path):
            logger.error(f"Config file not found: {self.config_path}")
            return False

        # Create test workspace and file
        test_file = self.create_test_workspace(file_path)

        # Detect expected language
        expected_language = self.test_language_detection(test_file)
        if not expected_language:
            logger.error("Could not determine expected language for file")
            self.cleanup()
            return False

        # Start MCP server with custom config
        config_abs_path = os.path.abspath(self.config_path)
        custom_cmd = ["go", "run", "main.go", "--config", config_abs_path]

        try:
            with MCPToolRunner(custom_cmd=custom_cmd) as runner:
                # Test LSP connection
                connection_success = self.test_lsp_connection(runner, expected_language)

                if connection_success:
                    # Test file analysis
                    analysis_success = self.test_file_analysis(runner, test_file)

                    if analysis_success:
                        logger.info("🎉 All tests passed!")
                        return True
                    else:
                        logger.error("❌ File analysis failed")
                        return False
                else:
                    logger.error("❌ LSP connection failed")
                    return False

        except Exception as e:
            logger.error(f"❌ Test runner failed: {e}")
            return False
        finally:
            self.cleanup()


def main():
    parser = argparse.ArgumentParser(
        description='Test LSP server initialization and connection',
        epilog='''Examples:

  # Test JavaScript file
  %(prog)s --config test_lsp_config.json --file test.js

  # Test TypeScript file with custom workspace
  %(prog)s --config test_lsp_config.json --file test.ts --workspace /tmp/my_test

  # Test Go file
  %(prog)s --config test_lsp_config.json --file test.go
'''
    )

    parser.add_argument('--config', required=True,
                       help='Path to LSP configuration file')
    parser.add_argument('--file', required=True,
                       help='Test file name or path (will be created in workspace)')
    parser.add_argument('--workspace',
                       help='Workspace directory path (default: /tmp/test_workspace)')

    args = parser.parse_args()

    try:
        tester = LSPInitTester(args.config, args.workspace)
        success = tester.run_test(args.file)
        sys.exit(0 if success else 1)

    except Exception as e:
        logger.error(f"Test failed with exception: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
