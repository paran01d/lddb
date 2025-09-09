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

	// Always scrape LDDB for the most up-to-date information
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

	// Check if we already have this UPC in our collection for reference
	existing, _ := h.dbService.GetLaserDiscByUPC(upc)
	
	// Prepare response with both LDDB result and local info
	response := gin.H{
		"source": "lddb.com",
		"result": result,
	}
	
	if existing != nil {
		response["existing"] = existing
		response["message"] = "LaserDisc found in LDDB (also exists in local collection)"
	} else {
		response["message"] = "LaserDisc information found in LDDB"
	}

	c.JSON(http.StatusOK, response)
}

// LookupByReference looks up LaserDisc information by catalog reference
// GET /api/lookup/reference/:reference
func (h *LookupHandler) LookupByReference(c *gin.Context) {
	reference := strings.TrimSpace(c.Param("reference"))
	if reference == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reference parameter is required"})
		return
	}

	// Scrape LDDB using reference search
	result, err := h.scraper.LookupByReference(reference)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to lookup LaserDisc information",
			"details": err.Error(),
		})
		return
	}

	if !result.Found {
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "LaserDisc not found",
			"reference": reference,
			"source":    "lddb.com",
			"error":     result.Error,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "LaserDisc information found by reference",
		"source":    "lddb.com",
		"reference": reference,
		"result":    result,
	})
}