package tools

import (
	"testing"
)

func TestApplyPagination(t *testing.T) {
	// Test data
	items := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	
	tests := []struct {
		name           string
		items          []string
		offset         int
		limit          int
		expectedCount  int
		expectedStart  int
		expectedEnd    int
		expectedHasMore bool
		expectedHasPrevious bool
	}{
		{
			name:           "Normal pagination",
			items:          items,
			offset:         0,
			limit:          3,
			expectedCount:  3,
			expectedStart:  1,
			expectedEnd:    3,
			expectedHasMore: true,
			expectedHasPrevious: false,
		},
		{
			name:           "Middle page",
			items:          items,
			offset:         3,
			limit:          3,
			expectedCount:  3,
			expectedStart:  4,
			expectedEnd:    6,
			expectedHasMore: true,
			expectedHasPrevious: true,
		},
		{
			name:           "Last page",
			items:          items,
			offset:         9,
			limit:          3,
			expectedCount:  1,
			expectedStart:  10,
			expectedEnd:    10,
			expectedHasMore: false,
			expectedHasPrevious: true,
		},
		{
			name:           "Offset exceeds total",
			items:          items,
			offset:         15,
			limit:          3,
			expectedCount:  0,
			expectedStart:  0,
			expectedEnd:    0,
			expectedHasMore: false,
			expectedHasPrevious: true,
		},
		{
			name:           "All items",
			items:          items,
			offset:         0,
			limit:          20,
			expectedCount:  10,
			expectedStart:  1,
			expectedEnd:    10,
			expectedHasMore: false,
			expectedHasPrevious: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paginated, result := ApplyPagination(tt.items, tt.offset, tt.limit)
			
			if len(paginated) != tt.expectedCount {
				t.Errorf("Expected %d items, got %d", tt.expectedCount, len(paginated))
			}
			
			if result.Start != tt.expectedStart {
				t.Errorf("Expected start %d, got %d", tt.expectedStart, result.Start)
			}
			
			if result.End != tt.expectedEnd {
				t.Errorf("Expected end %d, got %d", tt.expectedEnd, result.End)
			}
			
			if result.HasMore != tt.expectedHasMore {
				t.Errorf("Expected HasMore %v, got %v", tt.expectedHasMore, result.HasMore)
			}
			
			if result.HasPrevious != tt.expectedHasPrevious {
				t.Errorf("Expected HasPrevious %v, got %v", tt.expectedHasPrevious, result.HasPrevious)
			}
		})
	}
}

func TestFormatPaginationInfo(t *testing.T) {
	tests := []struct {
		name     string
		result   PaginationResult
		expected string
	}{
		{
			name: "Normal pagination",
			result: PaginationResult{
				Start:       1,
				End:         3,
				Total:       10,
				Count:       3,
				HasMore:     true,
				HasPrevious: false,
			},
			expected: "Showing results 1-3 of 10 total",
		},
		{
			name: "All results",
			result: PaginationResult{
				Start:       1,
				End:         10,
				Total:       10,
				Count:       10,
				HasMore:     false,
				HasPrevious: false,
			},
			expected: "Found 10 results",
		},
		{
			name: "No results",
			result: PaginationResult{
				Count:       0,
				Offset:      15,
				Total:       10,
				HasMore:     false,
				HasPrevious: true,
			},
			expected: "No results (offset 15 exceeds total 10)",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPaginationInfo(tt.result)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}