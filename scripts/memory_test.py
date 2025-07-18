#!/usr/bin/env python3
"""
Memory usage testing script for MCP-LSP Bridge analysis functionality.
Monitors memory consumption during various analysis operations.
"""

import resource
import time
import subprocess
import json
import sys
import os
from pathlib import Path

def get_memory_usage():
    """Get current memory usage in MB."""
    # Using resource module to get memory usage
    usage = resource.getrusage(resource.RUSAGE_SELF)
    # ru_maxrss is in kilobytes on Linux, bytes on macOS
    maxrss = usage.ru_maxrss
    # Convert to MB (assuming Linux where it's in KB)
    return maxrss / 1024 if maxrss > 100000 else maxrss / 1024 / 1024

def run_analysis_with_memory_monitoring(analysis_type, query, description):
    """Run analysis while monitoring memory usage."""
    print(f"\n{'='*60}")
    print(f"MEMORY TEST: {description}")
    print(f"Analysis Type: {analysis_type}")
    print(f"Query: {query}")
    print(f"{'='*60}")

    # Get baseline memory
    baseline_memory = get_memory_usage()
    print(f"Baseline memory: {baseline_memory:.1f} MB")

    cmd = [
        "uv", "run", "python", "scripts/test_mcp_tools.py",
        f"● project_analysis (MCP)(analysis_type=\"{analysis_type}\", query=\"{query}\")"
    ]

    start_time = time.time()
    peak_memory = baseline_memory

    try:
        # Start the process
        process = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)

        # Monitor memory usage during execution
        memory_samples = []
        while process.poll() is None:
            current_memory = get_memory_usage()
            memory_samples.append(current_memory)
            peak_memory = max(peak_memory, current_memory)
            time.sleep(0.1)  # Sample every 100ms

        # Get final result
        stdout, stderr = process.communicate()
        end_time = time.time()
        duration = end_time - start_time

        # Final memory measurement
        final_memory = get_memory_usage()
        memory_delta = final_memory - baseline_memory
        peak_delta = peak_memory - baseline_memory

        # Analyze result
        success = process.returncode == 0
        content_size = len(stdout) if stdout else 0

        print(f"✅ Status: {'SUCCESS' if success else 'FAILED'}")
        print(f"Duration: {duration:.2f}s")
        print(f"Final memory: {final_memory:.1f} MB")
        print(f"Memory delta: {memory_delta:+.1f} MB")
        print(f"Peak memory: {peak_memory:.1f} MB")
        print(f"Peak delta: {peak_delta:+.1f} MB")
        print(f"Response size: {content_size:,} bytes")

        if memory_samples:
            avg_memory = sum(memory_samples) / len(memory_samples)
            print(f"Average memory: {avg_memory:.1f} MB")

        # Memory efficiency metrics
        if content_size > 0:
            bytes_per_mb = content_size / max(peak_delta, 0.1)  # Avoid division by zero
            print(f"Efficiency: {bytes_per_mb:,.0f} bytes per MB of memory")

        return {
            'description': description,
            'analysis_type': analysis_type,
            'query': query,
            'success': success,
            'duration': duration,
            'baseline_memory': baseline_memory,
            'final_memory': final_memory,
            'peak_memory': peak_memory,
            'memory_delta': memory_delta,
            'peak_delta': peak_delta,
            'content_size': content_size,
            'avg_memory': avg_memory if memory_samples else baseline_memory
        }

    except Exception as e:
        print(f"💥 ERROR: {e}")
        return {
            'description': description,
            'analysis_type': analysis_type,
            'query': query,
            'success': False,
            'error': str(e)
        }

def main():
    """Run memory usage tests."""
    print("🧠 Starting MCP-LSP Bridge Memory Usage Tests")
    print(f"Timestamp: {time.strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"Python process PID: {os.getpid()}")

    # System memory info (simplified)
    try:
        with open('/proc/meminfo', 'r') as f:
            meminfo = f.read()
            for line in meminfo.split('\n'):
                if 'MemTotal:' in line:
                    total_kb = int(line.split()[1])
                    print(f"System total memory: {total_kb / 1024 / 1024:.1f} GB")
                elif 'MemAvailable:' in line:
                    available_kb = int(line.split()[1])
                    print(f"System available memory: {available_kb / 1024 / 1024:.1f} GB")
    except (FileNotFoundError, PermissionError):
        print("System memory info not available (not on Linux or no /proc/meminfo access)")

    # Test cases focused on memory usage
    test_cases = [
        # Simple tests
        ("file_analysis", "main.go", "Small file memory test"),
        ("pattern_analysis", "error_handling", "Simple pattern analysis"),

        # Complex tests
        ("file_analysis", "mcpserver/tools/project_analysis.go", "Large file memory test"),
        ("workspace_symbols", "ProjectAnalyzer", "Symbol search memory test"),
        ("references", "ProjectAnalyzer", "References memory test"),

        # Stress tests
        ("workspace_symbols", "*", "Wildcard symbol search"),
        ("text_search", "error", "Text search memory test"),
        ("document_symbols", "analysis/engine.go", "Complex document symbols"),

        # Error cases
        ("file_analysis", "nonexistent/file.go", "Error handling memory test"),
    ]

    results = []
    total_tests = len(test_cases)

    for i, (analysis_type, query, description) in enumerate(test_cases, 1):
        print(f"\n🧪 Memory Test {i}/{total_tests}")
        result = run_analysis_with_memory_monitoring(analysis_type, query, description)
        results.append(result)

        # Small delay between tests to allow memory cleanup
        time.sleep(1)

    # Analyze results
    print(f"\n{'='*80}")
    print("🧠 MEMORY USAGE ANALYSIS")
    print(f"{'='*80}")

    successful_results = [r for r in results if r.get('success', False)]

    if successful_results:
        # Memory statistics
        memory_deltas = [r['memory_delta'] for r in successful_results]
        peak_deltas = [r['peak_delta'] for r in successful_results]
        content_sizes = [r['content_size'] for r in successful_results]

        print(f"Successful tests: {len(successful_results)}/{total_tests}")
        print(f"Average memory delta: {sum(memory_deltas) / len(memory_deltas):+.1f} MB")
        print(f"Max memory delta: {max(memory_deltas):+.1f} MB")
        print(f"Average peak delta: {sum(peak_deltas) / len(peak_deltas):+.1f} MB")
        print(f"Max peak delta: {max(peak_deltas):+.1f} MB")
        print(f"Average response size: {sum(content_sizes) / len(content_sizes):,.0f} bytes")
        print(f"Max response size: {max(content_sizes):,.0f} bytes")

        # Memory efficiency
        total_content = sum(content_sizes)
        total_memory = sum(peak_deltas)
        if total_memory > 0:
            efficiency = total_content / total_memory
            print(f"Overall efficiency: {efficiency:,.0f} bytes per MB")

        # Memory concerns
        high_memory_tests = [r for r in successful_results if r['peak_delta'] > 50]  # >50MB
        if high_memory_tests:
            print(f"\n⚠️  High memory usage tests (>50MB):")
            for test in high_memory_tests:
                print(f"  {test['description']}: {test['peak_delta']:+.1f} MB peak")

        # Memory leaks check
        memory_leaks = [r for r in successful_results if r['memory_delta'] > 10]  # >10MB final delta
        if memory_leaks:
            print(f"\n🚨 Potential memory leaks (>10MB final delta):")
            for test in memory_leaks:
                print(f"  {test['description']}: {test['memory_delta']:+.1f} MB final delta")
        else:
            print(f"\n✅ No memory leaks detected (all final deltas <10MB)")

        # Performance vs memory analysis
        print(f"\n📊 Performance vs Memory Analysis:")
        for result in successful_results:
            if 'duration' in result:
                mb_per_second = result['peak_delta'] / result['duration'] if result['duration'] > 0 else 0
                print(f"  {result['description']}: {mb_per_second:.1f} MB/s")

    failed_results = [r for r in results if not r.get('success', False)]
    if failed_results:
        print(f"\n❌ Failed tests: {len(failed_results)}")
        for test in failed_results:
            print(f"  {test['description']}")

    print(f"\n🏁 Memory test completed!")

    # Return success if no major memory issues
    has_major_issues = any(r.get('peak_delta', 0) > 100 for r in successful_results)  # >100MB
    return not has_major_issues

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)
