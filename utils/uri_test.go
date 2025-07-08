package utils

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeURI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already normalized file URI",
			input:    "file:///home/user/file.go",
			expected: "file:///home/user/file.go",
		},
		{
			name:     "http URI unchanged",
			input:    "https://example.com/file",
			expected: "https://example.com/file",
		},
		{
			name:     "absolute path",
			input:    "/home/user/file.go",
			expected: "file:///home/user/file.go",
		},
		{
			name:     "relative path becomes absolute",
			input:    "file.go",
			expected: "file://" + mustAbs("file.go"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeURI(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeURI(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestURIToFilePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "file URI",
			input:    "file:///home/user/file.go",
			expected: "/home/user/file.go",
		},
		{
			name:     "already a file path",
			input:    "/home/user/file.go",
			expected: "/home/user/file.go",
		},
		{
			name:     "http URI unchanged",
			input:    "https://example.com/file",
			expected: "https://example.com/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := URIToFilePath(tt.input)
			if result != tt.expected {
				t.Errorf("URIToFilePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFilePathToURI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "absolute path",
			input:    "/home/user/file.go",
			expected: "file:///home/user/file.go",
		},
		{
			name:     "already a URI",
			input:    "file:///home/user/file.go",
			expected: "file:///home/user/file.go",
		},
		{
			name:     "relative path becomes absolute",
			input:    "file.go",
			expected: "file://" + mustAbs("file.go"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilePathToURI(tt.input)
			if result != tt.expected {
				t.Errorf("FilePathToURI(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	testPaths := []string{
		"/home/user/file.go",
		"/tmp/test.txt",
		"/var/log/app.log",
	}

	for _, path := range testPaths {
		t.Run(path, func(t *testing.T) {
			// Convert to URI and back
			uri := FilePathToURI(path)
			resultPath := URIToFilePath(uri)

			if resultPath != path {
				t.Errorf("Round trip failed: %s -> %s -> %s", path, uri, resultPath)
			}

			// Normalize the URI
			normalizedURI := NormalizeURI(path)
			if !strings.HasPrefix(normalizedURI, "file://") {
				t.Errorf("NormalizeURI(%s) = %s, should start with file://", path, normalizedURI)
			}
		})
	}
}

// mustAbs is a helper that calls filepath.Abs and panics on error (for tests only)
func mustAbs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return abs
}