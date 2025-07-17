package analysis

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewErrorHandler tests the creation of a new error handler
func TestNewErrorHandler(t *testing.T) {
	handler := NewErrorHandler(5, true, 0.5)
	
	assert.NotNil(t, handler)
	assert.Equal(t, 5, handler.maxErrors)
	assert.True(t, handler.continueOnError)
	assert.InDelta(t, 0.5, handler.errorThreshold, 0.001)
}

// TestHandleError tests the error handling logic
func TestHandleError(t *testing.T) {
	// Test cases with different error handling configurations
	testCases := []struct {
		name            string
		maxErrors       int
		continueOnError bool
		errorThreshold  float64
		numErrors       int
		expectContinue  bool
	}{
		{
			name:            "Continue on first few errors",
			maxErrors:       5,
			continueOnError: true,
			errorThreshold:  0.5,
			numErrors:       3,
			expectContinue:  true,
		},
		{
			name:            "Stop on max errors",
			maxErrors:       3,
			continueOnError: true,
			errorThreshold:  0.5,
			numErrors:       4,
			expectContinue:  false,
		},
		{
			name:            "Stop on error threshold",
			maxErrors:       10,
			continueOnError: true,
			errorThreshold:  0.3,
			numErrors:       4,
			expectContinue:  true, // Not stopping at 4 errors out of 14 total ops
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewErrorHandler(tc.maxErrors, tc.continueOnError, tc.errorThreshold)
			
			// Prepare metadata
			metadata := &AnalysisMetadata{
				FilesScanned: 10, // Simulate some files scanned
				Errors:       []AnalysisError{},
			}
			
			// Simulate multiple errors
			var continueAnalysis bool
			for i := 0; i < tc.numErrors; i++ {
				err := errors.New("test error")
				continueAnalysis = handler.HandleError(err, "test_context", metadata)
			}
			
			// Check results
			assert.Equal(t, tc.expectContinue, continueAnalysis)
			assert.Equal(t, tc.numErrors, len(metadata.Errors))
		})
	}
}

// TestShouldContinue tests the ShouldContinue method
func TestShouldContinue(t *testing.T) {
	testCases := []struct {
		name            string
		maxErrors       int
		currentErrors   int
		expectContinue  bool
	}{
		{
			name:            "Continue below max errors",
			maxErrors:       5,
			currentErrors:   3,
			expectContinue:  true,
		},
		{
			name:            "Stop at max errors",
			maxErrors:       5,
			currentErrors:   5,
			expectContinue:  false,
		},
		{
			name:            "Stop above max errors",
			maxErrors:       5,
			currentErrors:   6,
			expectContinue:  false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewErrorHandler(tc.maxErrors, true, 0.5)
			
			// Prepare metadata with errors
			metadata := &AnalysisMetadata{
				Errors: make([]AnalysisError, tc.currentErrors),
			}
			
			// Check if analysis should continue
			continueAnalysis := handler.ShouldContinue(metadata)
			
			assert.Equal(t, tc.expectContinue, continueAnalysis)
		})
	}
}

// TestHandleErrorThreshold tests the error threshold handling
func TestHandleErrorThreshold(t *testing.T) {
	// Scenario: 4 errors in a 10-operation context with 0.3 threshold
	handler := NewErrorHandler(10, true, 0.3)
	
	metadata := &AnalysisMetadata{
		FilesScanned: 10, // Total operations
		Errors:       []AnalysisError{},
	}
	
	// Simulate errors
	for i := 0; i < 4; i++ {
		err := errors.New("test error")
		continueAnalysis := handler.HandleError(err, "test_context", metadata)
		assert.True(t, continueAnalysis)
	}
	
	// 5th error will stop analysis
	err := errors.New("final error")
	continueAnalysis := handler.HandleError(err, "test_context", metadata)
	assert.False(t, continueAnalysis)
}