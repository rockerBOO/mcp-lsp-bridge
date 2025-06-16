// Package directories provides cross-platform directory resolution
// for applications based on user context and system conventions.
package directories

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

// DirectoryResolver handles directory resolution logic for applications
type DirectoryResolver struct {
	appName         string
	user            *user.User
	shouldEnsureDir bool
}

// NewDirectoryResolver creates a new directory resolver for the given application name
func NewDirectoryResolver(appName string, user *user.User, shouldEnsureDir bool) *DirectoryResolver {
	return &DirectoryResolver{appName: appName, user: user, shouldEnsureDir: shouldEnsureDir}
}

// isRoot checks if the current user is root (UID 0 on Unix systems)
func (dr *DirectoryResolver) isRoot(u *user.User) bool {
	return u.Uid == "0"
}

// maybeEnsureDir creates the directory if it doesn't exist and returns the path
func (dr *DirectoryResolver) maybeEnsureDir(dir string) (string, error) {
	if !dr.shouldEnsureDir {
		return dir, nil
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return dir, nil
}

// GetLogDirectory returns the appropriate log directory based on user context
// For root: /var/log/{appName}
// For regular users: ~/.local/share/{appName} (Unix) or %LOCALAPPDATA%\{appName}\logs (Windows)
func (dr *DirectoryResolver) GetLogDirectory() (string, error) {
	if dr.isRoot(dr.user) {
		return dr.maybeEnsureDir(filepath.Join("/", "var", "log", dr.appName))
	}

	return dr.getUserLogDirectory()
}

// getUserLogDirectory gets the user-specific log directory following platform conventions
func (dr *DirectoryResolver) getUserLogDirectory() (string, error) {
	if runtime.GOOS == "windows" {
		// Windows: use %LOCALAPPDATA%
		baseDir := os.Getenv("LOCALAPPDATA")
		if baseDir == "" {
			baseDir = filepath.Join(dr.user.HomeDir, "AppData", "Local")
		}
		return dr.maybeEnsureDir(filepath.Join(baseDir, dr.appName, "logs"))
	}

	// Unix-like systems: follow XDG Base Directory Specification
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome == "" {
		xdgDataHome = filepath.Join(dr.user.HomeDir, ".local", "share")
	}

	return dr.maybeEnsureDir(filepath.Join(xdgDataHome, dr.appName, "logs"))
}

// GetDataDirectory returns appropriate data directory for the user
// For root: /var/lib/{appName}
// For regular users: ~/.local/share/{appName} (Unix) or %LOCALAPPDATA%\{appName} (Windows)
func (dr *DirectoryResolver) GetDataDirectory() (string, error) {
	if dr.isRoot(dr.user) {
		return dr.maybeEnsureDir(filepath.Join("/", "var", "lib", dr.appName))
	}

	if runtime.GOOS == "windows" {
		baseDir := os.Getenv("LOCALAPPDATA")
		if baseDir == "" {
			baseDir = filepath.Join(dr.user.HomeDir, "AppData", "Local")
		}
		return dr.maybeEnsureDir(filepath.Join(baseDir, dr.appName))
	}

	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome == "" {
		xdgDataHome = filepath.Join(dr.user.HomeDir, ".local", "share")
	}

	return dr.maybeEnsureDir(filepath.Join(xdgDataHome, dr.appName))
}

// GetCacheDirectory returns appropriate cache directory for the user
// For root: /var/cache/{appName}
// For regular users: ~/.cache/{appName} (Unix) or %TEMP%\{appName} (Windows)
func (dr *DirectoryResolver) GetCacheDirectory() (string, error) {
	if dr.isRoot(dr.user) {
		return dr.maybeEnsureDir(filepath.Join("/", "var", "cache", dr.appName))
	}

	if runtime.GOOS == "windows" {
		baseDir := os.Getenv("TEMP")
		if baseDir == "" {
			baseDir = filepath.Join(dr.user.HomeDir, "AppData", "Local", "Temp")
		}
		return dr.maybeEnsureDir(filepath.Join(baseDir, dr.appName))
	}

	xdgCacheHome := os.Getenv("XDG_CACHE_HOME")
	if xdgCacheHome == "" {
		xdgCacheHome = filepath.Join(dr.user.HomeDir, ".cache")
	}

	return dr.maybeEnsureDir(filepath.Join(xdgCacheHome, dr.appName))
}

// GetConfigDirectory returns appropriate configuration directory for the user
// For root: /etc/{appName}
// For regular users: ~/.config/{appName} (Unix) or %APPDATA%\{appName} (Windows)
func (dr *DirectoryResolver) GetConfigDirectory() (string, error) {
	if dr.isRoot(dr.user) {
		return dr.maybeEnsureDir(filepath.Join("/", "etc", dr.appName))
	}

	if runtime.GOOS == "windows" {
		configDir := os.Getenv("APPDATA")
		if configDir == "" {
			configDir = filepath.Join(dr.user.HomeDir, "AppData", "Roaming")
		}
		return dr.maybeEnsureDir(filepath.Join(configDir, dr.appName))
	}

	// Unix-like systems
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		xdgConfigHome = filepath.Join(dr.user.HomeDir, ".config")
	}

	return dr.maybeEnsureDir(filepath.Join(xdgConfigHome, dr.appName))
}