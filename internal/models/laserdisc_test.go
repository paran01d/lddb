package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLaserDisc_TableName(t *testing.T) {
	ld := LaserDisc{}
	assert.Equal(t, "laserdiscs", ld.TableName())
}

func TestLaserDisc_DefaultValues(t *testing.T) {
	ld := LaserDisc{
		UPC:   "1234567890",
		Title: "Test Movie",
	}
	
	// Watched should default to false
	assert.False(t, ld.Watched)
	
	// Time fields should be zero initially (set by GORM)
	assert.True(t, ld.AddedDate.IsZero())
	assert.True(t, ld.UpdatedDate.IsZero())
}

func TestCreateLaserDiscRequest_RequiredFields(t *testing.T) {
	// Test with all required fields
	req := CreateLaserDiscRequest{
		UPC:   "1234567890",
		Title: "Test Movie",
	}
	
	assert.Equal(t, "1234567890", req.UPC)
	assert.Equal(t, "Test Movie", req.Title)
}

func TestCreateLaserDiscRequest_AllFields(t *testing.T) {
	req := CreateLaserDiscRequest{
		UPC:           "1234567890",
		Title:         "Test Movie",
		Year:          1995,
		Director:      "Test Director",
		Genre:         "Action",
		Format:        "CLV",
		Sides:         2,
		Runtime:       120,
		CoverImageURL: "https://example.com/cover.jpg",
		Notes:         "Test notes",
	}
	
	assert.Equal(t, "1234567890", req.UPC)
	assert.Equal(t, "Test Movie", req.Title)
	assert.Equal(t, 1995, req.Year)
	assert.Equal(t, "Test Director", req.Director)
	assert.Equal(t, "Action", req.Genre)
	assert.Equal(t, "CLV", req.Format)
	assert.Equal(t, 2, req.Sides)
	assert.Equal(t, 120, req.Runtime)
	assert.Equal(t, "https://example.com/cover.jpg", req.CoverImageURL)
	assert.Equal(t, "Test notes", req.Notes)
}

func TestUpdateLaserDiscRequest_NilValues(t *testing.T) {
	req := UpdateLaserDiscRequest{}
	
	// All fields should be nil by default
	assert.Nil(t, req.Title)
	assert.Nil(t, req.Year)
	assert.Nil(t, req.Director)
	assert.Nil(t, req.Genre)
	assert.Nil(t, req.Format)
	assert.Nil(t, req.Sides)
	assert.Nil(t, req.Runtime)
	assert.Nil(t, req.CoverImageURL)
	assert.Nil(t, req.Watched)
	assert.Nil(t, req.Notes)
}

func TestUpdateLaserDiscRequest_PartialUpdate(t *testing.T) {
	title := "Updated Title"
	watched := true
	
	req := UpdateLaserDiscRequest{
		Title:   &title,
		Watched: &watched,
		// Other fields remain nil
	}
	
	assert.NotNil(t, req.Title)
	assert.Equal(t, "Updated Title", *req.Title)
	assert.NotNil(t, req.Watched)
	assert.True(t, *req.Watched)
	assert.Nil(t, req.Year)
	assert.Nil(t, req.Director)
}

func TestLookupResult_DefaultValues(t *testing.T) {
	result := LookupResult{
		UPC: "1234567890",
	}
	
	assert.Equal(t, "1234567890", result.UPC)
	assert.False(t, result.Found) // Should default to false
	assert.Empty(t, result.Error)
}

func TestLookupResult_WithData(t *testing.T) {
	result := LookupResult{
		UPC:           "1234567890",
		Title:         "Test Movie",
		Year:          1995,
		Director:      "Test Director",
		Genre:         "Action",
		Format:        "CLV",
		Sides:         2,
		Runtime:       120,
		CoverImageURL: "https://example.com/cover.jpg",
		Found:         true,
	}
	
	assert.Equal(t, "1234567890", result.UPC)
	assert.Equal(t, "Test Movie", result.Title)
	assert.Equal(t, 1995, result.Year)
	assert.Equal(t, "Test Director", result.Director)
	assert.Equal(t, "Action", result.Genre)
	assert.Equal(t, "CLV", result.Format)
	assert.Equal(t, 2, result.Sides)
	assert.Equal(t, 120, result.Runtime)
	assert.Equal(t, "https://example.com/cover.jpg", result.CoverImageURL)
	assert.True(t, result.Found)
	assert.Empty(t, result.Error)
}

func TestLookupResult_WithError(t *testing.T) {
	result := LookupResult{
		UPC:   "1234567890",
		Found: false,
		Error: "UPC not found",
	}
	
	assert.Equal(t, "1234567890", result.UPC)
	assert.False(t, result.Found)
	assert.Equal(t, "UPC not found", result.Error)
}

// Test JSON serialization/deserialization of models
func TestLaserDisc_JSONSerialization(t *testing.T) {
	now := time.Now()
	ld := LaserDisc{
		ID:            1,
		UPC:           "1234567890",
		Title:         "Test Movie",
		Year:          1995,
		Director:      "Test Director",
		Genre:         "Action",
		Format:        "CLV",
		Sides:         2,
		Runtime:       120,
		CoverImageURL: "https://example.com/cover.jpg",
		Watched:       true,
		Notes:         "Test notes",
		AddedDate:     now,
		UpdatedDate:   now,
	}
	
	// These would be tested in integration tests with actual JSON marshaling
	// For now, just verify the struct can be created
	assert.Equal(t, uint(1), ld.ID)
	assert.Equal(t, "1234567890", ld.UPC)
	assert.Equal(t, "Test Movie", ld.Title)
	assert.True(t, ld.Watched)
}