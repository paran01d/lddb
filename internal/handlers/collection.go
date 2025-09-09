package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/paran01d/lddb/internal/database"
)

// CollectionHandler handles collection-related HTTP requests
type CollectionHandler struct {
	dbService *database.Service
}

// NewCollectionHandler creates a new collection handler
func NewCollectionHandler(dbService *database.Service) *CollectionHandler {
	return &CollectionHandler{
		dbService: dbService,
	}
}

// GetCollection retrieves all LaserDiscs in the collection
func (h *CollectionHandler) GetCollection(c *gin.Context) {
	// TODO: Implement
	c.JSON(http.StatusOK, gin.H{"message": "Collection endpoint - TODO"})
}

// AddLaserDisc adds a new LaserDisc to the collection
func (h *CollectionHandler) AddLaserDisc(c *gin.Context) {
	// TODO: Implement
	c.JSON(http.StatusOK, gin.H{"message": "Add LaserDisc endpoint - TODO"})
}

// UpdateLaserDisc updates an existing LaserDisc
func (h *CollectionHandler) UpdateLaserDisc(c *gin.Context) {
	// TODO: Implement
	c.JSON(http.StatusOK, gin.H{"message": "Update LaserDisc endpoint - TODO"})
}

// DeleteLaserDisc deletes a LaserDisc from the collection
func (h *CollectionHandler) DeleteLaserDisc(c *gin.Context) {
	// TODO: Implement
	c.JSON(http.StatusOK, gin.H{"message": "Delete LaserDisc endpoint - TODO"})
}

// ToggleWatched toggles the watched status of a LaserDisc
func (h *CollectionHandler) ToggleWatched(c *gin.Context) {
	// TODO: Implement
	c.JSON(http.StatusOK, gin.H{"message": "Toggle watched endpoint - TODO"})
}

// GetRandomUnwatched returns a random unwatched LaserDisc
func (h *CollectionHandler) GetRandomUnwatched(c *gin.Context) {
	// TODO: Implement
	c.JSON(http.StatusOK, gin.H{"message": "Random unwatched endpoint - TODO"})
}