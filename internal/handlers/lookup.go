package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/paran01d/lddb/internal/database"
)

// LookupHandler handles lookup-related HTTP requests
type LookupHandler struct {
	dbService *database.Service
}

// NewLookupHandler creates a new lookup handler
func NewLookupHandler(dbService *database.Service) *LookupHandler {
	return &LookupHandler{
		dbService: dbService,
	}
}

// LookupByUPC looks up LaserDisc information by UPC
func (h *LookupHandler) LookupByUPC(c *gin.Context) {
	// TODO: Implement
	c.JSON(http.StatusOK, gin.H{"message": "Lookup by UPC endpoint - TODO"})
}