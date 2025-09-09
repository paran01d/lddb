package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/paran01d/lddb/internal/models"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *Service {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.LaserDisc{})
	require.NoError(t, err)

	return NewService(db)
}

// createTestLaserDisc creates a test LaserDisc
func createTestLaserDisc() *models.CreateLaserDiscRequest {
	return &models.CreateLaserDiscRequest{
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
}

func TestService_CreateLaserDisc(t *testing.T) {
	service := setupTestDB(t)
	req := createTestLaserDisc()

	// Test creating a LaserDisc
	laserdisc, err := service.CreateLaserDisc(req)
	require.NoError(t, err)
	assert.NotZero(t, laserdisc.ID)
	assert.Equal(t, req.UPC, laserdisc.UPC)
	assert.Equal(t, req.Title, laserdisc.Title)
	assert.Equal(t, req.Year, laserdisc.Year)
	assert.Equal(t, req.Director, laserdisc.Director)
	assert.Equal(t, req.Genre, laserdisc.Genre)
	assert.Equal(t, req.Format, laserdisc.Format)
	assert.Equal(t, req.Sides, laserdisc.Sides)
	assert.Equal(t, req.Runtime, laserdisc.Runtime)
	assert.Equal(t, req.CoverImageURL, laserdisc.CoverImageURL)
	assert.Equal(t, req.Notes, laserdisc.Notes)
	assert.False(t, laserdisc.Watched) // Should default to false
	assert.NotZero(t, laserdisc.AddedDate)
}

func TestService_CreateLaserDisc_DuplicateUPC(t *testing.T) {
	service := setupTestDB(t)
	req := createTestLaserDisc()

	// Create first LaserDisc
	_, err := service.CreateLaserDisc(req)
	require.NoError(t, err)

	// Try to create another with same UPC
	_, err = service.CreateLaserDisc(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestService_GetLaserDiscByID(t *testing.T) {
	service := setupTestDB(t)
	req := createTestLaserDisc()

	// Create a LaserDisc
	created, err := service.CreateLaserDisc(req)
	require.NoError(t, err)

	// Get it by ID
	retrieved, err := service.GetLaserDiscByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.Title, retrieved.Title)

	// Test non-existent ID
	_, err = service.GetLaserDiscByID(999999)
	assert.Error(t, err)
}

func TestService_GetLaserDiscByUPC(t *testing.T) {
	service := setupTestDB(t)
	req := createTestLaserDisc()

	// Create a LaserDisc
	created, err := service.CreateLaserDisc(req)
	require.NoError(t, err)

	// Get it by UPC
	retrieved, err := service.GetLaserDiscByUPC(created.UPC)
	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.UPC, retrieved.UPC)

	// Test non-existent UPC
	_, err = service.GetLaserDiscByUPC("nonexistent")
	assert.Error(t, err)
}

func TestService_GetAllLaserDiscs(t *testing.T) {
	service := setupTestDB(t)

	// Initially should be empty
	laserdiscs, err := service.GetAllLaserDiscs()
	require.NoError(t, err)
	assert.Empty(t, laserdiscs)

	// Create some test LaserDiscs
	req1 := createTestLaserDisc()
	req1.UPC = "1111111111"
	req1.Title = "Movie A"

	req2 := createTestLaserDisc()
	req2.UPC = "2222222222"
	req2.Title = "Movie B"

	_, err = service.CreateLaserDisc(req1)
	require.NoError(t, err)
	_, err = service.CreateLaserDisc(req2)
	require.NoError(t, err)

	// Get all
	laserdiscs, err = service.GetAllLaserDiscs()
	require.NoError(t, err)
	assert.Len(t, laserdiscs, 2)
	// Should be ordered by title
	assert.Equal(t, "Movie A", laserdiscs[0].Title)
	assert.Equal(t, "Movie B", laserdiscs[1].Title)
}

func TestService_UpdateLaserDisc(t *testing.T) {
	service := setupTestDB(t)
	req := createTestLaserDisc()

	// Create a LaserDisc
	created, err := service.CreateLaserDisc(req)
	require.NoError(t, err)

	// Update it
	newTitle := "Updated Title"
	newYear := 2000
	watched := true
	updateReq := &models.UpdateLaserDiscRequest{
		Title:   &newTitle,
		Year:    &newYear,
		Watched: &watched,
	}

	updated, err := service.UpdateLaserDisc(created.ID, updateReq)
	require.NoError(t, err)
	assert.Equal(t, newTitle, updated.Title)
	assert.Equal(t, newYear, updated.Year)
	assert.True(t, updated.Watched)
	// Other fields should remain unchanged
	assert.Equal(t, created.Director, updated.Director)
	assert.Equal(t, created.Genre, updated.Genre)

	// Test updating non-existent ID
	_, err = service.UpdateLaserDisc(999999, updateReq)
	assert.Error(t, err)
}

func TestService_DeleteLaserDisc(t *testing.T) {
	service := setupTestDB(t)
	req := createTestLaserDisc()

	// Create a LaserDisc
	created, err := service.CreateLaserDisc(req)
	require.NoError(t, err)

	// Delete it
	err = service.DeleteLaserDisc(created.ID)
	require.NoError(t, err)

	// Should no longer exist
	_, err = service.GetLaserDiscByID(created.ID)
	assert.Error(t, err)

	// Test deleting non-existent ID
	err = service.DeleteLaserDisc(999999)
	assert.Error(t, err)
}

func TestService_ToggleWatched(t *testing.T) {
	service := setupTestDB(t)
	req := createTestLaserDisc()

	// Create a LaserDisc (default watched = false)
	created, err := service.CreateLaserDisc(req)
	require.NoError(t, err)
	assert.False(t, created.Watched)

	// Toggle to watched
	toggled, err := service.ToggleWatched(created.ID)
	require.NoError(t, err)
	assert.True(t, toggled.Watched)

	// Toggle back to unwatched
	toggled, err = service.ToggleWatched(created.ID)
	require.NoError(t, err)
	assert.False(t, toggled.Watched)

	// Test toggling non-existent ID
	_, err = service.ToggleWatched(999999)
	assert.Error(t, err)
}

func TestService_GetRandomUnwatched(t *testing.T) {
	service := setupTestDB(t)

	// Test when no unwatched discs exist
	_, err := service.GetRandomUnwatched()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no unwatched")

	// Create some test LaserDiscs
	req1 := createTestLaserDisc()
	req1.UPC = "1111111111"
	req1.Title = "Unwatched Movie 1"

	req2 := createTestLaserDisc()
	req2.UPC = "2222222222"
	req2.Title = "Unwatched Movie 2"

	req3 := createTestLaserDisc()
	req3.UPC = "3333333333"
	req3.Title = "Watched Movie"

	disc1, err := service.CreateLaserDisc(req1)
	require.NoError(t, err)
	disc2, err := service.CreateLaserDisc(req2)
	require.NoError(t, err)
	disc3, err := service.CreateLaserDisc(req3)
	require.NoError(t, err)

	// Mark one as watched
	_, err = service.ToggleWatched(disc3.ID)
	require.NoError(t, err)

	// Get random unwatched should return one of the unwatched ones
	random, err := service.GetRandomUnwatched()
	require.NoError(t, err)
	assert.False(t, random.Watched)
	assert.True(t, random.ID == disc1.ID || random.ID == disc2.ID)
}

func TestService_SearchLaserDiscs(t *testing.T) {
	service := setupTestDB(t)

	// Create some test LaserDiscs
	req1 := createTestLaserDisc()
	req1.UPC = "1111111111"
	req1.Title = "Star Wars"
	req1.Director = "George Lucas"
	req1.Genre = "Sci-Fi"

	req2 := createTestLaserDisc()
	req2.UPC = "2222222222"
	req2.Title = "Star Trek"
	req2.Director = "J.J. Abrams"
	req2.Genre = "Sci-Fi"

	req3 := createTestLaserDisc()
	req3.UPC = "3333333333"
	req3.Title = "The Matrix"
	req3.Director = "Wachowski Sisters"
	req3.Genre = "Action"

	_, err := service.CreateLaserDisc(req1)
	require.NoError(t, err)
	_, err = service.CreateLaserDisc(req2)
	require.NoError(t, err)
	_, err = service.CreateLaserDisc(req3)
	require.NoError(t, err)

	// Search by title
	results, err := service.SearchLaserDiscs("Star")
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Search by director
	results, err = service.SearchLaserDiscs("Lucas")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Star Wars", results[0].Title)

	// Search by genre
	results, err = service.SearchLaserDiscs("Sci-Fi")
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Search with no matches
	results, err = service.SearchLaserDiscs("NonExistent")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestService_GetStats(t *testing.T) {
	service := setupTestDB(t)

	// Initially should be all zeros
	stats, err := service.GetStats()
	require.NoError(t, err)
	assert.Equal(t, int64(0), stats["total"])
	assert.Equal(t, int64(0), stats["watched"])
	assert.Equal(t, int64(0), stats["unwatched"])

	// Create some test LaserDiscs
	req1 := createTestLaserDisc()
	req1.UPC = "1111111111"
	req2 := createTestLaserDisc()
	req2.UPC = "2222222222"
	req3 := createTestLaserDisc()
	req3.UPC = "3333333333"

	disc1, err := service.CreateLaserDisc(req1)
	require.NoError(t, err)
	disc2, err := service.CreateLaserDisc(req2)
	require.NoError(t, err)
	_, err = service.CreateLaserDisc(req3)
	require.NoError(t, err)

	// Mark two as watched
	_, err = service.ToggleWatched(disc1.ID)
	require.NoError(t, err)
	_, err = service.ToggleWatched(disc2.ID)
	require.NoError(t, err)

	// Check stats
	stats, err = service.GetStats()
	require.NoError(t, err)
	assert.Equal(t, int64(3), stats["total"])
	assert.Equal(t, int64(2), stats["watched"])
	assert.Equal(t, int64(1), stats["unwatched"])
}