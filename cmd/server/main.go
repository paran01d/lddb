package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/paran01d/lddb/internal/database"
	"github.com/paran01d/lddb/internal/handlers"
	"github.com/paran01d/lddb/internal/models"
)

func main() {
	// Initialize database in data directory
	db, err := gorm.Open(sqlite.Open("data/collection.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.LaserDisc{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize database service
	dbService := database.NewService(db)

	// Initialize handlers
	collectionHandler := handlers.NewCollectionHandler(dbService)
	lookupHandler := handlers.NewLookupHandler(dbService)

	// Initialize Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})

	// Serve static files
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")

	// Web routes
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "LaserDisc Collection Manager",
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Collection endpoints
		api.GET("/collection", collectionHandler.GetCollection)
		api.POST("/collection", collectionHandler.AddLaserDisc)
		api.PUT("/collection/:id", collectionHandler.UpdateLaserDisc)
		api.DELETE("/collection/:id", collectionHandler.DeleteLaserDisc)
		api.POST("/collection/:id/watched", collectionHandler.ToggleWatched)

		// Lookup and random endpoints
		api.GET("/lookup/:upc", lookupHandler.LookupByUPC)
		api.GET("/lookup/reference/:reference", lookupHandler.LookupByReference)
		api.GET("/random-unwatched", collectionHandler.GetRandomUnwatched)
	}

	// Find an available port starting from 8080
	port := findAvailablePort(8080)
	log.Printf("Starting server on :%d", port)
	log.Fatal(router.Run(fmt.Sprintf(":%d", port)))
}

// findAvailablePort finds an available port starting from the given port number
func findAvailablePort(startPort int) int {
	for port := startPort; port < startPort+100; port++ {
		if isPortAvailable(port) {
			return port
		}
	}
	// If no port found in range, fall back to original
	return startPort
}

// isPortAvailable checks if a port is available for binding
func isPortAvailable(port int) bool {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	defer listener.Close()
	return true
}