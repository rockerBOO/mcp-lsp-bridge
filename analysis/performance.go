package analysis

import (
	"runtime"
	"time"
)

// PerformanceConfig defines performance-related settings for analysis
type PerformanceConfig struct {
	MaxGoroutines    int
	Timeout          time.Duration
	MemoryLimit      int64 // bytes
	EnableProfiling  bool
	BatchSize        int
}

// DefaultPerformanceConfig creates a default configuration optimized for most systems
func DefaultPerformanceConfig() *PerformanceConfig {
	return &PerformanceConfig{
		MaxGoroutines:   runtime.NumCPU() * 2,
		Timeout:         30 * time.Second,
		MemoryLimit:     1024 * 1024 * 1024, // 1GB
		EnableProfiling: false,
		BatchSize:       50,
	}
}

