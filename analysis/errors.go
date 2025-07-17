package analysis

import (
	"strings"
)

// AnalysisErrorHandler manages errors during project analysis
type AnalysisErrorHandler struct {
	maxErrors      int
	continueOnError bool
	errorThreshold  float64 // % of failed operations before stopping
}

// NewErrorHandler creates a new error handler with specified configurations
func NewErrorHandler(maxErrors int, continueOnError bool, threshold float64) *AnalysisErrorHandler {
	return &AnalysisErrorHandler{
		maxErrors:      maxErrors,
		continueOnError: continueOnError,
		errorThreshold:  threshold,
	}
}

// HandleError processes an error during analysis
func (h *AnalysisErrorHandler) HandleError(err error, context string, metadata *AnalysisMetadata) bool {
	analysisErr := AnalysisError{
		Message: err.Error(),
		Type:    "error",
	}
	
	// Extract language from context if possible
	if strings.Contains(context, "language:") {
		// TODO: Implement language parsing from context
		// This is a placeholder for future implementation
		_ = context
	}
	
	metadata.Errors = append(metadata.Errors, analysisErr)
	
	// Check if we should continue
	if len(metadata.Errors) >= h.maxErrors {
		return false
	}
	
	// Check error threshold
	totalOps := metadata.FilesScanned + len(metadata.Errors)
	if totalOps > 0 {
		errorRate := float64(len(metadata.Errors)) / float64(totalOps)
		if errorRate > h.errorThreshold {
			return false
		}
	}
	
	return h.continueOnError
}

// ShouldContinue checks if analysis should continue based on error conditions
func (h *AnalysisErrorHandler) ShouldContinue(metadata *AnalysisMetadata) bool {
	return len(metadata.Errors) < h.maxErrors
}