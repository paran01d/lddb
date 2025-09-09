package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/paran01d/lddb/internal/database"
	"github.com/paran01d/lddb/internal/models"
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
// GET /api/collection?search=query&limit=10&offset=0
func (h *CollectionHandler) GetCollection(c *gin.Context) {
	// Get query parameters
	search := c.Query("search")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter (1-100)"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
		return
	}

	var laserdiscs []models.LaserDisc
	var total int64

	if search != "" {
		// Search with query
		laserdiscs, err = h.dbService.SearchLaserDiscs(search)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search collection"})
			return
		}
		total = int64(len(laserdiscs))
		
		// Apply pagination to search results
		if offset >= len(laserdiscs) {
			laserdiscs = []models.LaserDisc{}
		} else {
			end := offset + limit
			if end > len(laserdiscs) {
				end = len(laserdiscs)
			}
			laserdiscs = laserdiscs[offset:end]
		}
	} else {
		// Get all with pagination
		laserdiscs, err = h.dbService.GetAllLaserDiscs()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve collection"})
			return
		}
		total = int64(len(laserdiscs))

		// Apply pagination
		if offset >= len(laserdiscs) {
			laserdiscs = []models.LaserDisc{}
		} else {
			end := offset + limit
			if end > len(laserdiscs) {
				end = len(laserdiscs)
			}
			laserdiscs = laserdiscs[offset:end]
		}
	}

	// Get collection statistics
	stats, err := h.dbService.GetStats()
	if err != nil {
		stats = map[string]interface{}{
			"total":     0,
			"watched":   0,
			"unwatched": 0,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"laserdiscs": laserdiscs,
		"pagination": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
		"stats": stats,
	})
}

// AddLaserDisc adds a new LaserDisc to the collection
// POST /api/collection
func (h *CollectionHandler) AddLaserDisc(c *gin.Context) {
	var req models.CreateLaserDiscRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	laserdisc, err := h.dbService.CreateLaserDisc(&req)
	if err != nil {
		if err.Error() == "laserdisc with this UPC already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create LaserDisc", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "LaserDisc added successfully",
		"laserdisc": laserdisc,
	})
}

// UpdateLaserDisc updates an existing LaserDisc
// PUT /api/collection/:id
func (h *CollectionHandler) UpdateLaserDisc(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid LaserDisc ID"})
		return
	}

	var req models.UpdateLaserDiscRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	laserdisc, err := h.dbService.UpdateLaserDisc(uint(id), &req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "LaserDisc not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update LaserDisc", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "LaserDisc updated successfully",
		"laserdisc": laserdisc,
	})
}

// DeleteLaserDisc deletes a LaserDisc from the collection
// DELETE /api/collection/:id
func (h *CollectionHandler) DeleteLaserDisc(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid LaserDisc ID"})
		return
	}

	err = h.dbService.DeleteLaserDisc(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "LaserDisc not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete LaserDisc", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "LaserDisc deleted successfully"})
}

// ToggleWatched toggles the watched status of a LaserDisc
// POST /api/collection/:id/watched
func (h *CollectionHandler) ToggleWatched(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid LaserDisc ID"})
		return
	}

	laserdisc, err := h.dbService.ToggleWatched(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "LaserDisc not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update watched status", "details": err.Error()})
		return
	}

	status := "unwatched"
	if laserdisc.Watched {
		status = "watched"
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Watched status updated successfully",
		"status":     status,
		"laserdisc": laserdisc,
	})
}

// GetRandomUnwatched returns a random unwatched LaserDisc
// GET /api/random-unwatched
func (h *CollectionHandler) GetRandomUnwatched(c *gin.Context) {
	laserdisc, err := h.dbService.GetRandomUnwatched()
	if err != nil {
		if err.Error() == "no unwatched laserdiscs found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "No unwatched LaserDiscs found in collection"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get random unwatched LaserDisc", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Random unwatched LaserDisc selected",
		"laserdisc": laserdisc,
	})
}