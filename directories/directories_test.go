package directories

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNewDirectoryResolver(t *testing.T) {
	tests := []struct {
		name            string
		appName         string
		user            *user.User
		shouldEnsureDir bool
		want            *DirectoryResolver
	}{
		{
			name:    "basic constructor",
			appName: "testapp",
			user: &user.User{
				Uid:     "1000",
				HomeDir: "/home/testuser",
			},
			shouldEnsureDir: true,
			want: &DirectoryResolver{
				appName:         "testapp",
				user:            &user.User{Uid: "1000", HomeDir: "/home/testuser"},
				shouldEnsureDir: true,
			},
		},
		{
			name:    "without ensuring directories",
			appName: "myapp",
			user: &user.User{
				Uid:     "0",
				HomeDir: "/root",
			},
			shouldEnsureDir: false,
			want: &DirectoryResolver{
				appName:         "myapp",
				user:            &user.User{Uid: "0", HomeDir: "/root"},
				shouldEnsureDir: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDirectoryResolver(tt.appName, tt.user, tt.shouldEnsureDir)
			if got.appName != tt.want.appName {
				t.Errorf("NewDirectoryResolver() appName = %v, want %v", got.appName, tt.want.appName)
			}
			if got.shouldEnsureDir != tt.want.shouldEnsureDir {
				t.Errorf("NewDirectoryResolver() shouldEnsureDir = %v, want %v", got.shouldEnsureDir, tt.want.shouldEnsureDir)
			}
			if got.user.Uid != tt.want.user.Uid {
				t.Errorf("NewDirectoryResolver() user.Uid = %v, want %v", got.user.Uid, tt.want.user.Uid)
			}
		})
	}
}

func TestDirectoryResolver_isRoot(t *testing.T) {
	dr := &DirectoryResolver{}

	tests := []struct {
		name string
		user *user.User
		want bool
	}{
		{
			name: "root user",
			user: &user.User{Uid: "0"},
			want: true,
		},
		{
			name: "regular user",
			user: &user.User{Uid: "1000"},
			want: false,
		},
		{
			name: "another regular user",
			user: &user.User{Uid: "501"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dr.isRoot(tt.user)
			if got != tt.want {
				t.Errorf("isRoot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDirectoryResolver_maybeEnsureDir(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name            string
		shouldEnsureDir bool
		dir             string
		wantErr         bool
	}{
		{
			name:            "ensure directory - success",
			shouldEnsureDir: true,
			dir:             filepath.Join(tempDir, "test", "subdir"),
			wantErr:         false,
		},
		{
			name:            "don't ensure directory",
			shouldEnsureDir: false,
			dir:             filepath.Join(tempDir, "nonexistent"),
			wantErr:         false,
		},
		{
			name:            "ensure directory - invalid path",
			shouldEnsureDir: true,
			dir:             "/root/invalid/path/for/test",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := &DirectoryResolver{shouldEnsureDir: tt.shouldEnsureDir}
			got, err := dr.maybeEnsureDir(tt.dir)

			if (err != nil) != tt.wantErr {
				t.Errorf("maybeEnsureDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got != tt.dir {
					t.Errorf("maybeEnsureDir() = %v, want %v", got, tt.dir)
				}

				// If we should ensure the directory and no error occurred, verify it exists
				if tt.shouldEnsureDir {
					if _, statErr := os.Stat(tt.dir); os.IsNotExist(statErr) {
						t.Errorf("maybeEnsureDir() should have created directory %v", tt.dir)
					}
				}
			}
		})
	}
}

func TestDirectoryResolver_GetLogDirectory(t *testing.T) {
	tests := []struct {
		name            string
		appName         string
		u               *user.User
		shouldEnsureDir bool
		want            string
		wantErr         bool
	}{
		{
			name:    "root",
			appName: "test",
			u: &user.User{
				Uid: "0",
			},
			shouldEnsureDir: false,
			want:            "/var/log/test",
			wantErr:         false,
		},
		{
			name:    "regular user",
			appName: "test",
			u: &user.User{
				Uid:     "1000",
				HomeDir: "/home/testuser",
			},
			shouldEnsureDir: false,
			want:            filepath.Join("/home/testuser", ".local", "share", "test", "logs"),
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := NewDirectoryResolver(tt.appName, tt.u, tt.shouldEnsureDir)
			got, gotErr := dr.GetLogDirectory()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetLogDirectory() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetLogDirectory() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("GetLogDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDirectoryResolver_getUserLogDirectory(t *testing.T) {
	// Note: We can't actually change runtime.GOOS in tests, but we can test the logic

	tests := []struct {
		name            string
		appName         string
		user            *user.User
		shouldEnsureDir bool
		envVars         map[string]string
		wantContains    string // What the result should contain
		wantErr         bool
	}{
		{
			name:    "unix user with XDG_DATA_HOME",
			appName: "testapp",
			user: &user.User{
				Uid:     "1000",
				HomeDir: "/home/testuser",
			},
			shouldEnsureDir: false,
			envVars: map[string]string{
				"XDG_DATA_HOME": "/custom/data",
			},
			wantContains: "testapp/logs",
			wantErr:      false,
		},
		{
			name:    "unix user without XDG_DATA_HOME",
			appName: "testapp",
			user: &user.User{
				Uid:     "1000",
				HomeDir: "/home/testuser",
			},
			shouldEnsureDir: false,
			envVars:         map[string]string{},
			wantContains:    ".local/share/testapp/logs",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				oldValue := os.Getenv(key)
				os.Setenv(key, value)
				defer os.Setenv(key, oldValue)
			}

			// Clear environment variables not in the test case
			if _, exists := tt.envVars["XDG_DATA_HOME"]; !exists {
				oldValue := os.Getenv("XDG_DATA_HOME")
				os.Unsetenv("XDG_DATA_HOME")
				defer os.Setenv("XDG_DATA_HOME", oldValue)
			}

			dr := NewDirectoryResolver(tt.appName, tt.user, tt.shouldEnsureDir)
			got, err := dr.getUserLogDirectory()

			if (err != nil) != tt.wantErr {
				t.Errorf("getUserLogDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !strings.Contains(got, tt.wantContains) {
				t.Errorf("getUserLogDirectory() = %v, should contain %v", got, tt.wantContains)
			}
		})
	}
}

func TestDirectoryResolver_GetDataDirectory(t *testing.T) {
	tests := []struct {
		name            string
		appName         string
		user            *user.User
		shouldEnsureDir bool
		want            string
		wantErr         bool
	}{
		{
			name:    "root",
			appName: "test",
			user: &user.User{
				Uid: "0",
			},
			shouldEnsureDir: false,
			want:            "/var/lib/test",
			wantErr:         false,
		},
		{
			name:    "regular user",
			appName: "test",
			user: &user.User{
				Uid:     "1000",
				HomeDir: "/home/testuser",
			},
			shouldEnsureDir: false,
			want:            filepath.Join("/home/testuser", ".local", "share", "test"),
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := NewDirectoryResolver(tt.appName, tt.user, tt.shouldEnsureDir)
			got, gotErr := dr.GetDataDirectory()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetDataDirectory() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetDataDirectory() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("GetDataDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDirectoryResolver_GetCacheDirectory(t *testing.T) {
	tests := []struct {
		name            string
		appName         string
		user            *user.User
		shouldEnsureDir bool
		want            string
		wantErr         bool
	}{
		{
			name:    "root",
			appName: "test",
			user: &user.User{
				Uid: "0",
			},
			shouldEnsureDir: false,
			want:            "/var/cache/test",
			wantErr:         false,
		},
		{
			name:    "regular user",
			appName: "test",
			user: &user.User{
				Uid:     "1000",
				HomeDir: "/home/testuser",
			},
			shouldEnsureDir: false,
			want:            filepath.Join("/home/testuser", ".cache", "test"),
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := NewDirectoryResolver(tt.appName, tt.user, tt.shouldEnsureDir)
			got, gotErr := dr.GetCacheDirectory()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetCacheDirectory() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetCacheDirectory() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("GetCacheDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDirectoryResolver_GetConfigDirectory(t *testing.T) {
	tests := []struct {
		name            string
		appName         string
		user            *user.User
		shouldEnsureDir bool
		want            string
		wantErr         bool
	}{
		{
			name:    "root",
			appName: "test",
			user: &user.User{
				Uid: "0",
			},
			shouldEnsureDir: false,
			want:            "/etc/test",
			wantErr:         false,
		},
		{
			name:    "regular user",
			appName: "test",
			user: &user.User{
				Uid:     "1000",
				HomeDir: "/home/testuser",
			},
			shouldEnsureDir: false,
			want:            filepath.Join("/home/testuser", ".config", "test"),
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := NewDirectoryResolver(tt.appName, tt.user, tt.shouldEnsureDir)
			got, gotErr := dr.GetConfigDirectory()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetConfigDirectory() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetConfigDirectory() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("GetConfigDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDirectoryResolver_WithEnvironmentVariables(t *testing.T) {
	// Test environment variable handling for Unix systems
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix environment variable tests on Windows")
	}

	tests := []struct {
		name     string
		envVars  map[string]string
		function string // which function to test
		contains string // what the result should contain
	}{
		{
			name: "XDG_CONFIG_HOME override",
			envVars: map[string]string{
				"XDG_CONFIG_HOME": "/custom/config",
			},
			function: "config",
			contains: "/custom/config/testapp",
		},
		{
			name: "XDG_CACHE_HOME override",
			envVars: map[string]string{
				"XDG_CACHE_HOME": "/custom/cache",
			},
			function: "cache",
			contains: "/custom/cache/testapp",
		},
		{
			name: "XDG_DATA_HOME override",
			envVars: map[string]string{
				"XDG_DATA_HOME": "/custom/data",
			},
			function: "data",
			contains: "/custom/data/testapp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalEnv := make(map[string]string)
			for key := range tt.envVars {
				originalEnv[key] = os.Getenv(key)
			}

			// Set test environment
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Restore environment after test
			defer func() {
				for key, value := range originalEnv {
					if value == "" {
						os.Unsetenv(key)
					} else {
						os.Setenv(key, value)
					}
				}
			}()

			user := &user.User{
				Uid:     "1000",
				HomeDir: "/home/testuser",
			}
			dr := NewDirectoryResolver("testapp", user, false)

			var got string
			var err error

			switch tt.function {
			case "config":
				got, err = dr.GetConfigDirectory()
			case "cache":
				got, err = dr.GetCacheDirectory()
			case "data":
				got, err = dr.GetDataDirectory()
			default:
				t.Fatalf("Unknown function: %s", tt.function)
			}

			if err != nil {
				t.Errorf("Function failed: %v", err)
				return
			}

			if !strings.Contains(got, tt.contains) {
				t.Errorf("Result %v should contain %v", got, tt.contains)
			}
		})
	}
}

func TestDirectoryResolver_DirectoryEnsuring(t *testing.T) {
	tests := []struct {
		name            string
		shouldEnsureDir bool
		function        string
		wantDirExists   bool
	}{
		{
			name:            "ensure directories enabled",
			shouldEnsureDir: true,
			function:        "config",
			wantDirExists:   true,
		},
		{
			name:            "ensure directories disabled",
			shouldEnsureDir: false,
			function:        "config",
			wantDirExists:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh temp directory for each test case
			tempDir := t.TempDir()
			
			testUser := &user.User{
				Uid:     "1000",
				HomeDir: tempDir,
			}

			dr := NewDirectoryResolver("testapp", testUser, tt.shouldEnsureDir)

			var got string
			var err error

			switch tt.function {
			case "config":
				got, err = dr.GetConfigDirectory()
			case "cache":
				got, err = dr.GetCacheDirectory()
			case "data":
				got, err = dr.GetDataDirectory()
			case "log":
				got, err = dr.GetLogDirectory()
			}

			if err != nil {
				t.Errorf("Function failed: %v", err)
				return
			}

			// Check if directory exists
			_, statErr := os.Stat(got)
			dirExists := !os.IsNotExist(statErr)

			if dirExists != tt.wantDirExists {
				t.Errorf("Directory existence = %v, want %v (path: %v)", dirExists, tt.wantDirExists, got)
			}
		})
	}
}

func TestDirectoryResolver_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		appName string
		user    *user.User
		wantErr bool
	}{
		{
			name:    "empty app name",
			appName: "",
			user: &user.User{
				Uid:     "1000",
				HomeDir: "/home/test",
			},
			wantErr: false, // Should work, just create empty path segment
		},
		{
			name:    "app name with special characters",
			appName: "my-app_v2.0",
			user: &user.User{
				Uid:     "1000",
				HomeDir: "/home/test",
			},
			wantErr: false,
		},
		{
			name:    "user with empty home directory",
			appName: "testapp",
			user: &user.User{
				Uid:     "1000",
				HomeDir: "",
			},
			wantErr: false, // Should still work, might result in relative paths
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr := NewDirectoryResolver(tt.appName, tt.user, false)

			functions := []func() (string, error){
				dr.GetConfigDirectory,
				dr.GetCacheDirectory,
				dr.GetDataDirectory,
				dr.GetLogDirectory,
			}

			for i, fn := range functions {
				_, err := fn()
				if (err != nil) != tt.wantErr {
					t.Errorf("Function %d error = %v, wantErr %v", i, err, tt.wantErr)
				}
			}
		})
	}
}