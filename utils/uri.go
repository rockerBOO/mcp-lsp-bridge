package utils

import (
	"path/filepath"
	"strings"
)

// NormalizeURI ensures the URI has the proper file:// scheme
func NormalizeURI(uri string) string {
	// If it already has a file scheme, return as-is
	if strings.HasPrefix(uri, "file://") {
		return uri
	}

	// If it has any other scheme (http://, https://, etc.), return as-is
	if strings.Contains(uri, "://") {
		return uri
	}

	// If it's an absolute path, convert to file URI
	if strings.HasPrefix(uri, "/") {
		return "file://" + uri
	}

	// If it's a relative path, convert to absolute path first, then to file URI
	if absPath, err := filepath.Abs(uri); err == nil {
		return "file://" + absPath
	}

	// Fallback: assume it's a file path and add file:// prefix
	return "file://" + uri
}

// URIToFilePath converts a file URI to a local file path
func URIToFilePath(uri string) string {
	return strings.TrimPrefix(uri, "file://")
}

// FilePathToURI converts a local file path to a file URI
func FilePathToURI(path string) string {
	if strings.HasPrefix(path, "file://") {
		return path // Already a URI
	}

	// Convert to absolute path if relative
	if absPath, err := filepath.Abs(path); err == nil {
		path = absPath
	}

	return "file://" + path
}
