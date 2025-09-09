package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/paran01d/lddb/internal/database"
	"github.com/paran01d/lddb/internal/scraper"
)

// LookupHandler handles lookup-related HTTP requests
type LookupHandler struct {
	dbService *database.Service
	scraper   *scraper.LDDBScraper
}

// NewLookupHandler creates a new lookup handler
func NewLookupHandler(dbService *database.Service) *LookupHandler {
	return &LookupHandler{
		dbService: dbService,
		scraper:   scraper.NewLDDBScraper(),
	}
}

// LookupByUPC looks up LaserDisc information by UPC
// GET /api/lookup/:upc
func (h *LookupHandler) LookupByUPC(c *gin.Context) {
	upc := strings.TrimSpace(c.Param("upc"))
	if upc == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UPC parameter is required"})
		return
	}

	// First check if we already have this UPC in our collection
	existing, err := h.dbService.GetLaserDiscByUPC(upc)
	if err == nil && existing != nil {
		c.JSON(http.StatusOK, gin.H{
			"message":     "LaserDisc found in local collection",
			"source":      "local",
			"laserdisc":  existing,
		})
		return
	}

	// If not found locally, scrape LDDB
	result, err := h.scraper.LookupByUPC(upc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to lookup LaserDisc information",
			"details": err.Error(),
		})
		return
	}

	if !result.Found {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "LaserDisc not found",
			"upc":     upc,
			"source":  "lddb.com",
			"error":   result.Error,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "LaserDisc information found",
		"source":  "lddb.com",
		"result":  result,
	})
}