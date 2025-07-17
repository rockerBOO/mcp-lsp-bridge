#!/usr/bin/env python3
"""
Performance testing script for project analysis functionality.
Tests various analysis types with different file sizes and complexity.
"""

import time
import sys
import json
import subprocess
from pathlib import Path

def run_analysis_test(analysis_type, query, description):
    """Run a single analysis test and measure performance."""
    print(f"\n{'='*60}")
    print(f"TEST: {description}")
    print(f"Analysis Type: {analysis_type}")
    print(f"Query: {query}")
    print(f"{'='*60}")
    
    cmd = [
        "uv", "run", "python", "scripts/test_mcp_tools.py",
        f"● project_analysis (MCP)(analysis_type=\"{analysis_type}\", query=\"{query}\")"
    ]
    
    start_time = time.time()
    
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=60)
        end_time = time.time()
        duration = end_time - start_time
        
        if result.returncode == 0:
            # Parse the JSON response
            lines = result.stdout.strip().split('\n')
            for line in lines:
                if line.startswith('{'):
                    try:
                        response = json.loads(line)
                        if 'content' in response and response['content']:
                            content = response['content'][0]['text']
                            # Extract key metrics from the response
                            lines_count = content.count('\n')
                            print(f"✅ SUCCESS")
                            print(f"Duration: {duration:.2f}s")
                            print(f"Response length: {len(content)} chars, {lines_count} lines")
                            
                            # Extract specific metrics if available
                            if "Complexity Score:" in content:
                                for line in content.split('\n'):
                                    if "Complexity Score:" in line:
                                        print(f"Complexity: {line.strip()}")
                            
                            return duration, True, content
                    except json.JSONDecodeError:
                        continue
            
            print(f"✅ SUCCESS (no JSON found)")
            print(f"Duration: {duration:.2f}s")
            return duration, True, result.stdout
        else:
            print(f"❌ FAILED")
            print(f"Duration: {duration:.2f}s")
            print(f"Error: {result.stderr}")
            return duration, False, result.stderr
            
    except subprocess.TimeoutExpired:
        print(f"⏰ TIMEOUT (>60s)")
        return 60.0, False, "Timeout"
    except Exception as e:
        print(f"💥 ERROR: {e}")
        return 0, False, str(e)

def main():
    """Run comprehensive performance tests."""
    print("🚀 Starting MCP-LSP Bridge Performance Tests")
    print(f"Timestamp: {time.strftime('%Y-%m-%d %H:%M:%S')}")
    
    # Test cases: (analysis_type, query, description)
    test_cases = [
        # File analysis tests
        ("file_analysis", "main.go", "Small file analysis (main.go)"),
        ("file_analysis", "mcpserver/tools/project_analysis.go", "Large file analysis (project_analysis.go)"),
        ("file_analysis", "analysis/engine.go", "Complex file analysis (analysis/engine.go)"),
        ("file_analysis", "scripts/test_mcp_tools.py", "Multi-language file analysis (Python)"),
        
        # Pattern analysis tests  
        ("pattern_analysis", "error_handling", "Error handling pattern analysis"),
        ("pattern_analysis", "naming_conventions", "Naming conventions analysis"),
        ("pattern_analysis", "architecture_patterns", "Architecture patterns analysis"),
        
        # Workspace symbol tests
        ("workspace_symbols", "LanguageClient", "Symbol search (LanguageClient)"),
        ("workspace_symbols", "ProjectAnalyzer", "Symbol search (ProjectAnalyzer)"),
        ("workspace_symbols", "Handler", "Symbol search (Handler)"),
        
        # Document symbol tests
        ("document_symbols", "mcpserver/tools/project_analysis.go", "Document symbols (large file)"),
        ("document_symbols", "main.go", "Document symbols (small file)"),
        
        # References tests
        ("references", "ProjectAnalyzer", "References analysis"),
        ("references", "HandleError", "References analysis (error handling)"),
        
        # Definitions tests
        ("definitions", "NewProjectAnalyzer", "Definitions analysis"),
        
        # Text search tests
        ("text_search", "TODO", "Text search (TODO)"),
        ("text_search", "error", "Text search (error)"),
        
        # Error handling tests
        ("file_analysis", "nonexistent/file.go", "Error handling (non-existent file)"),
        ("pattern_analysis", "invalid_pattern", "Error handling (invalid pattern)"),
    ]
    
    results = []
    total_tests = len(test_cases)
    passed_tests = 0
    
    for i, (analysis_type, query, description) in enumerate(test_cases, 1):
        print(f"\n📊 Test {i}/{total_tests}")
        duration, success, content = run_analysis_test(analysis_type, query, description)
        
        results.append({
            'test_number': i,
            'analysis_type': analysis_type,
            'query': query,
            'description': description,
            'duration': duration,
            'success': success,
            'content_length': len(content) if content else 0
        })
        
        if success:
            passed_tests += 1
    
    # Print summary
    print(f"\n{'='*80}")
    print("📈 PERFORMANCE TEST SUMMARY")
    print(f"{'='*80}")
    print(f"Total tests: {total_tests}")
    print(f"Passed: {passed_tests}")
    print(f"Failed: {total_tests - passed_tests}")
    print(f"Success rate: {(passed_tests / total_tests) * 100:.1f}%")
    
    # Performance statistics
    successful_durations = [r['duration'] for r in results if r['success']]
    if successful_durations:
        avg_duration = sum(successful_durations) / len(successful_durations)
        min_duration = min(successful_durations)
        max_duration = max(successful_durations)
        
        print(f"\nPerformance Statistics:")
        print(f"Average duration: {avg_duration:.2f}s")
        print(f"Fastest test: {min_duration:.2f}s")
        print(f"Slowest test: {max_duration:.2f}s")
    
    # Detailed results by analysis type
    print(f"\n📋 Results by Analysis Type:")
    for analysis_type in set(r['analysis_type'] for r in results):
        type_results = [r for r in results if r['analysis_type'] == analysis_type]
        type_successes = [r for r in type_results if r['success']]
        type_avg = sum(r['duration'] for r in type_successes) / len(type_successes) if type_successes else 0
        
        print(f"  {analysis_type}: {len(type_successes)}/{len(type_results)} passed, avg {type_avg:.2f}s")
    
    # Performance concerns
    slow_tests = [r for r in results if r['success'] and r['duration'] > 10]
    if slow_tests:
        print(f"\n⚠️  Slow tests (>10s):")
        for test in slow_tests:
            print(f"  {test['description']}: {test['duration']:.2f}s")
    
    failed_tests = [r for r in results if not r['success']]
    if failed_tests:
        print(f"\n❌ Failed tests:")
        for test in failed_tests:
            print(f"  {test['description']}: {test['duration']:.2f}s")
    
    print(f"\n🏁 Performance test completed!")
    return passed_tests == total_tests

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)