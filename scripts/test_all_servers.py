#!/usr/bin/env python3
"""
Comprehensive test for all language servers in lsp_config.json
Tests each server connection and reports which ones are available vs missing.
"""

import argparse
import json
import logging
import os
import subprocess
import sys
import tempfile
import time
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, List, Optional, Tuple

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s: %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)

@dataclass
class TestResult:
    server_name: str
    languages: List[str]
    command: str
    available: bool
    connection_success: bool
    error_message: Optional[str] = None
    response_time: Optional[float] = None

class LanguageServerTester:
    def __init__(self, config_path: str, mcp_command: str):
        self.config_path = config_path
        self.mcp_command = mcp_command
        self.results: List[TestResult] = []

        # Load configuration
        with open(config_path, 'r') as f:
            self.config = json.load(f)

        # File extensions for testing each language
        self.test_files = {
            'go': 'package main\n\nfunc main() {}\n',
            'python': 'def hello():\n    print("world")\n',
            'typescript': 'const x: number = 42;\n',
            'javascript': 'const x = 42;\n',
            'rust': 'fn main() {\n    println!("Hello, world!");\n}\n',
            'java': 'public class Test {\n    public static void main(String[] args) {}\n}\n',
            'cpp': '#include <iostream>\nint main() { return 0; }\n',
            'c': '#include <stdio.h>\nint main() { return 0; }\n',
            'csharp': 'using System;\nclass Program {\n    static void Main() {}\n}\n',
            'lua': 'print("Hello, world!")\n',
            'ruby': 'puts "Hello, world!"\n',
            'php': '<?php\necho "Hello, world!";\n?>\n',
            'swift': 'print("Hello, world!")\n',
            'kotlin': 'fun main() {\n    println("Hello, world!")\n}\n',
            'scala': 'object Main {\n  def main(args: Array[String]): Unit = {}\n}\n',
            'haskell': 'main :: IO ()\nmain = putStrLn "Hello, world!"\n',
            'elm': 'module Main exposing (main)\nmain = text "Hello"\n',
            'ocaml': 'let () = print_endline "Hello, world!"\n',
            'zig': 'const std = @import("std");\npub fn main() void {}\n',
            'dockerfile': 'FROM alpine:latest\nRUN echo "hello"\n',
            'yaml': 'name: test\nversion: 1.0\n',
            'json': '{"name": "test", "version": "1.0"}\n',
            'html': '<!DOCTYPE html>\n<html><body><h1>Test</h1></body></html>\n',
            'css': 'body {\n  margin: 0;\n  padding: 0;\n}\n'
        }

        # File extensions mapping
        self.extensions = {
            'go': '.go',
            'python': '.py',
            'typescript': '.ts',
            'javascript': '.js',
            'rust': '.rs',
            'java': '.java',
            'cpp': '.cpp',
            'c': '.c',
            'csharp': '.cs',
            'lua': '.lua',
            'ruby': '.rb',
            'php': '.php',
            'swift': '.swift',
            'kotlin': '.kt',
            'scala': '.scala',
            'haskell': '.hs',
            'elm': '.elm',
            'ocaml': '.ml',
            'zig': '.zig',
            'dockerfile': '.dockerfile',
            'yaml': '.yaml',
            'json': '.json',
            'html': '.html',
            'css': '.css'
        }

    def check_command_available(self, command: str) -> bool:
        """Check if a command is available in PATH"""
        try:
            subprocess.run(['which', command],
                         capture_output=True,
                         check=True,
                         timeout=5)
            return True
        except (subprocess.CalledProcessError, subprocess.TimeoutExpired, FileNotFoundError):
            return False

    def test_language_connection(self, language: str) -> Tuple[bool, Optional[str], Optional[float]]:
        """Test LSP connection for a specific language"""
        try:
            start_time = time.time()

            # Use the existing test_mcp_tools.py script
            cmd = [
                'uv', 'run', 'python', 'scripts/test_mcp_tools.py',
                f'lsp_connect (MCP)(language="{language}")',
                '--cmd', self.mcp_command
            ]

            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=30,
                cwd=Path(__file__).parent.parent
            )

            response_time = time.time() - start_time

            if result.returncode == 0:
                # Check if the response indicates success
                output_lines = result.stdout.strip().split('\n')

                # Find the JSON block (could be multiline)
                json_start = -1
                for i, line in enumerate(output_lines):
                    if line.strip().startswith('{'):
                        json_start = i
                        break

                if json_start >= 0:
                    # Find the end of JSON block
                    json_lines = []
                    brace_count = 0
                    for i in range(json_start, len(output_lines)):
                        line = output_lines[i].strip()
                        if not line:
                            continue
                        json_lines.append(line)

                        # Count braces to find end of JSON
                        brace_count += line.count('{') - line.count('}')
                        if brace_count == 0 and line.endswith('}'):
                            break

                    json_text = ''.join(json_lines)
                    try:
                        output = json.loads(json_text)
                        if 'content' in output and len(output['content']) > 0:
                            content = output['content'][0].get('text', '')
                            if f'Connected to LSP for {language}' in content:
                                return True, None, response_time
                            else:
                                return False, f'Unexpected response: {content}', response_time
                    except json.JSONDecodeError as e:
                        return False, f'Invalid JSON response: {str(e)}', response_time
                else:
                    return False, f'No JSON response found in output: {result.stdout[:200]}', response_time
            else:
                # Parse error from stderr if available
                error_msg = result.stderr.strip() if result.stderr else result.stdout.strip()
                return False, error_msg, response_time

        except subprocess.TimeoutExpired:
            return False, 'Connection timeout (30s)', 30.0
        except Exception as e:
            return False, f'Test error: {str(e)}', None

    def run_comprehensive_test(self):
        """Run comprehensive test for all configured servers"""
        logging.info("=== Starting Comprehensive Language Server Test ===")
        logging.info(f"Configuration: {self.config_path}")
        logging.info(f"MCP Command: {self.mcp_command}")

        language_servers = self.config.get('language_servers', {})
        language_server_map = self.config.get('language_server_map', {})

        logging.info(f"Found {len(language_servers)} language servers configured")
        logging.info(f"Found {len(language_server_map)} server-to-language mappings")

        for server_name, languages in language_server_map.items():
            if server_name not in language_servers:
                logging.warning(f"Server '{server_name}' in language_server_map but not in language_servers")
                continue

            server_config = language_servers[server_name]
            command = server_config.get('command', '')

            logging.info(f"\n--- Testing {server_name} ---")
            logging.info(f"Command: {command}")
            logging.info(f"Languages: {', '.join(languages)}")

            # Check if command is available
            available = self.check_command_available(command)
            logging.info(f"Command available: {'✅' if available else '❌'}")

            if not available:
                result = TestResult(
                    server_name=server_name,
                    languages=languages,
                    command=command,
                    available=False,
                    connection_success=False,
                    error_message=f"Command '{command}' not found in PATH"
                )
                self.results.append(result)
                continue

            # Test connection for each language
            for language in languages:
                logging.info(f"Testing connection for {language}...")
                success, error, response_time = self.test_language_connection(language)

                if success:
                    logging.info(f"✅ {language}: Connected successfully ({response_time:.2f}s)")
                else:
                    time_str = f"{response_time:.2f}s" if response_time else "N/A"
                    logging.error(f"❌ {language}: {error} ({time_str})")

                result = TestResult(
                    server_name=server_name,
                    languages=[language],
                    command=command,
                    available=available,
                    connection_success=success,
                    error_message=error,
                    response_time=response_time
                )
                self.results.append(result)

    def print_summary(self):
        """Print comprehensive test summary"""
        logging.info("\n" + "="*60)
        logging.info("COMPREHENSIVE TEST SUMMARY")
        logging.info("="*60)

        # Group results by server
        server_results = {}
        for result in self.results:
            if result.server_name not in server_results:
                server_results[result.server_name] = []
            server_results[result.server_name].append(result)

        total_servers = len(server_results)
        available_servers = sum(1 for results in server_results.values()
                              if any(r.available for r in results))
        working_servers = sum(1 for results in server_results.values()
                            if any(r.connection_success for r in results))

        logging.info(f"📊 OVERVIEW:")
        logging.info(f"   Total servers tested: {total_servers}")
        logging.info(f"   Available servers: {available_servers}")
        logging.info(f"   Working servers: {working_servers}")
        logging.info(f"   Success rate: {(working_servers/total_servers*100):.1f}%")

        # Detailed results
        logging.info(f"\n📋 DETAILED RESULTS:")

        for server_name, results in sorted(server_results.items()):
            first_result = results[0]
            languages = []
            for r in results:
                languages.extend(r.languages)

            status = "🟢" if any(r.connection_success for r in results) else \
                    "🟡" if first_result.available else "🔴"

            logging.info(f"   {status} {server_name}")
            logging.info(f"      Command: {first_result.command}")
            logging.info(f"      Languages: {', '.join(languages)}")

            if first_result.available:
                successful_langs = [r.languages[0] for r in results if r.connection_success]
                failed_langs = [r.languages[0] for r in results if not r.connection_success]

                if successful_langs:
                    logging.info(f"      ✅ Working: {', '.join(successful_langs)}")
                if failed_langs:
                    logging.info(f"      ❌ Failed: {', '.join(failed_langs)}")

                # Show first error for failed connections
                for r in results:
                    if not r.connection_success and r.error_message:
                        logging.info(f"      Error: {r.error_message}")
                        break
            else:
                logging.info(f"      ❌ Not installed: {first_result.error_message}")

        # Installation suggestions
        missing_servers = [results[0] for results in server_results.values()
                         if not any(r.available for r in results)]

        if missing_servers:
            logging.info(f"\n💡 INSTALLATION SUGGESTIONS:")
            for result in missing_servers:
                command = result.command
                suggestions = {
                    'gopls': 'go install golang.org/x/tools/gopls@latest',
                    'pyright-langserver': 'npm install -g pyright',
                    'typescript-language-server': 'npm install -g typescript-language-server',
                    'rust-analyzer': 'rustup component add rust-analyzer',
                    'clangd': 'sudo apt-get install clangd (Ubuntu/Debian)',
                    'lua-language-server': 'Install from: https://github.com/LuaLS/lua-language-server',
                }

                suggestion = suggestions.get(command, f'Install {command} according to its documentation')
                logging.info(f"   {command}: {suggestion}")

def main():
    parser = argparse.ArgumentParser(description='Test all language servers in configuration')
    parser.add_argument('--config', default='lsp_config.example.json',
                       help='Path to LSP configuration file')
    parser.add_argument('--cmd', default='go run main.go --config lsp_config.example.json',
                       help='MCP server command to use for testing')
    parser.add_argument('--timeout', type=int, default=30,
                       help='Timeout for each test in seconds')

    args = parser.parse_args()

    if not os.path.exists(args.config):
        logging.error(f"Configuration file not found: {args.config}")
        sys.exit(1)

    tester = LanguageServerTester(args.config, args.cmd)
    tester.run_comprehensive_test()
    tester.print_summary()

if __name__ == '__main__':
    main()
