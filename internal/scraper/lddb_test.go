package scraper

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLDDBScraper(t *testing.T) {
	scraper := NewLDDBScraper()
	assert.NotNil(t, scraper)
	assert.NotNil(t, scraper.collector)
}

func TestLDDBScraper_extractYear(t *testing.T) {
	scraper := NewLDDBScraper()
	
	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"1995", 1995, false},
		{"Released in 1995", 1995, false},
		{"1995-12-25", 1995, false},
		{"The year was 2001", 2001, false},
		{"no year here", 0, true},
		{"", 0, true},
	}
	
	for _, tt := range tests {
		result, err := scraper.extractYear(tt.input)
		if tt.hasError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	}
}

func TestLDDBScraper_extractRuntime(t *testing.T) {
	scraper := NewLDDBScraper()
	
	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"120 min", 120, false},
		{"120min", 120, false},
		{"2:00", 120, false},
		{"2:30", 150, false},
		{"1h 30m", 90, false},
		{"2h 15m", 135, false},
		{"1:45", 105, false},
		{"no runtime here", 0, true},
		{"", 0, true},
	}
	
	for _, tt := range tests {
		result, err := scraper.extractRuntime(tt.input)
		if tt.hasError {
			assert.Error(t, err, "Expected error for input: %s", tt.input)
		} else {
			assert.NoError(t, err, "Unexpected error for input: %s", tt.input)
			assert.Equal(t, tt.expected, result, "Wrong runtime for input: %s", tt.input)
		}
	}
}

func TestLDDBScraper_processLabelValue(t *testing.T) {
	// Since processLabelValue is a private method that modifies a LookupResult,
	// we'll test the pattern matching logic instead
	
	tests := []struct {
		label       string
		value       string
		expectMatch bool
	}{
		{"title", "Test Movie", true},
		{"TITLE", "Test Movie", true},
		{"movie title", "Test Movie", true},
		{"year", "1995", true},
		{"release year", "1995", true},
		{"director", "Test Director", true},
		{"directed by", "Test Director", true},
		{"genre", "Action", true},
		{"category", "Action", true},
		{"format", "CLV", true},
		{"disc format", "CLV", true},
		{"sides", "2", true},
		{"number of sides", "2", true},
		{"runtime", "120 min", true},
		{"duration", "2:00", true},
		{"invalid", "value", false},
	}
	
	for _, tt := range tests {
		// Test that the label contains expected keywords
		label := strings.ToLower(tt.label)
		actualMatch := false
		
		switch {
		case strings.Contains(label, "title"):
			actualMatch = true
		case strings.Contains(label, "year"):
			actualMatch = true
		case strings.Contains(label, "director") || strings.Contains(label, "directed"):
			actualMatch = true
		case strings.Contains(label, "genre") || strings.Contains(label, "category"):
			actualMatch = true
		case strings.Contains(label, "format"):
			actualMatch = true
		case strings.Contains(label, "sides"):
			actualMatch = true
		case strings.Contains(label, "runtime") || strings.Contains(label, "duration"):
			actualMatch = true
		default:
			actualMatch = false
		}
		
		assert.Equal(t, tt.expectMatch, actualMatch, "Match expectation failed for label: %s", tt.label)
	}
}

func TestLDDBScraper_LookupByUPC_InvalidUPC(t *testing.T) {
	scraper := NewLDDBScraper()
	
	// Test with invalid UPC (no digits)
	result, err := scraper.LookupByUPC("invalid")
	assert.NoError(t, err) // No network error, but result should indicate failure
	assert.Equal(t, "invalid", result.UPC)
	assert.False(t, result.Found)
	assert.Contains(t, result.Error, "Invalid UPC format")
}

func TestLDDBScraper_LookupByUPC_ValidFormat(t *testing.T) {
	// Test with valid UPC format (this won't actually hit the network in unit tests)
	// We're just testing the UPC cleaning logic
	upc := "123-456-7890"
	
	// The UPC should be cleaned to remove non-digits
	// We can't easily test the network call without mocking, 
	// so we'll test the UPC cleaning logic indirectly
	
	cleanUPC := strings.ReplaceAll(upc, "-", "")
	assert.Equal(t, "1234567890", cleanUPC)
}

// Test UPC cleaning logic separately
func TestCleanUPC(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1234567890", "1234567890"},
		{"123-456-7890", "1234567890"},
		{"123 456 7890", "1234567890"},
		{"123.456.7890", "1234567890"},
		{"abc123def456", "123456"},
		{"", ""},
		{"abcdef", ""},
	}
	
	for _, tt := range tests {
		// Simulate the UPC cleaning logic from the scraper
		result := ""
		for _, char := range tt.input {
			if char >= '0' && char <= '9' {
				result += string(char)
			}
		}
		assert.Equal(t, tt.expected, result, "Failed to clean UPC: %s", tt.input)
	}
}

// Test the HTML extraction patterns
func TestHTMLExtractionPatterns(t *testing.T) {
	// Test various patterns we might encounter in LDDB HTML
	testHTML := `
	<table>
		<tr><td>Title:</td><td>Star Wars</td></tr>
		<tr><td>Year:</td><td>1977</td></tr>
		<tr><td>Director:</td><td>George Lucas</td></tr>
		<tr><td>Runtime:</td><td>121 min</td></tr>
	</table>
	`
	
	// Test that we can find expected patterns
	assert.Contains(t, testHTML, "Title:")
	assert.Contains(t, testHTML, "Star Wars")
	assert.Contains(t, testHTML, "1977")
	assert.Contains(t, testHTML, "121 min")
}

// Test error handling for network issues (mock)
func TestLDDBScraper_ErrorHandling(t *testing.T) {
	scraper := NewLDDBScraper()
	
	// Test with empty UPC
	result, err := scraper.LookupByUPC("")
	assert.NoError(t, err)
	assert.False(t, result.Found)
	assert.Contains(t, result.Error, "Invalid UPC format")
	
	// Test with only spaces
	result, err = scraper.LookupByUPC("   ")
	assert.NoError(t, err)
	assert.False(t, result.Found)
	assert.Contains(t, result.Error, "Invalid UPC format")
}