package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/paran01d/lddb/internal/database"
	"github.com/paran01d/lddb/internal/handlers"
	"github.com/paran01d/lddb/internal/models"
)

// Global access token (generated on startup)
var accessToken string

func main() {
	// Generate mobile-friendly access token
	accessToken = generateMobileToken()
	log.Printf("🔑 Access Token: %s", accessToken)
	log.Printf("   Enter this token when prompted to access the application")
	
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

	// Add authentication middleware
	router.Use(authMiddleware())

	// Serve static files
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")

	// Web routes
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "LaserDisc Collection Manager",
		})
	})
	
	// Token authentication page
	router.GET("/auth", func(c *gin.Context) {
		c.HTML(http.StatusOK, "auth.html", gin.H{
			"title": "Access Required - LaserDisc Collection Manager",
		})
	})
	
	// Token validation endpoint
	router.POST("/auth/validate", func(c *gin.Context) {
		var req struct {
			Token string `json:"token" binding:"required"`
		}
		
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}
		
		token := strings.TrimSpace(strings.ToUpper(req.Token))
		if token == accessToken {
			c.JSON(http.StatusOK, gin.H{
				"message": "Access granted",
				"token":   accessToken,
			})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
		}
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

// generateMobileToken creates a mobile-friendly access token
func generateMobileToken() string {
	words := []string{
		"MOON", "STAR", "FIRE", "WAVE", "ROCK", "TREE", "BIRD", "FISH",
		"BLUE", "GOLD", "NEON", "JADE", "RUBY", "CYAN", "MINT", "ROSE",
		"COOL", "WARM", "FAST", "SLOW", "HIGH", "DEEP", "WILD", "CALM",
		"EPIC", "MEGA", "NOVA", "ZERO", "HERO", "SAGE", "FURY", "GLOW",
	}
	
	// Pick two random words
	word1 := words[randomInt(len(words))]
	word2 := words[randomInt(len(words))]
	
	// Generate 4-digit number
	number := randomInt(9000) + 1000
	
	return fmt.Sprintf("%s-%s-%d", word1, word2, number)
}

// randomInt generates a cryptographically secure random integer
func randomInt(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		log.Printf("Warning: Failed to generate secure random number, using fallback")
		return 0
	}
	return int(n.Int64())
}

// authMiddleware checks for valid access token
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for token entry page, validation endpoint, and static files
		path := c.Request.URL.Path
		if path == "/auth" || path == "/auth/validate" || strings.HasPrefix(path, "/static/") {
			c.Next()
			return
		}
		
		// Check for token in header or query parameter
		token := c.GetHeader("Authorization")
		if token == "" {
			token = c.Query("token")
		}
		
		// Remove "Bearer " prefix if present
		token = strings.TrimPrefix(token, "Bearer ")
		
		if token != accessToken {
			// Redirect to token entry page
			if c.GetHeader("Accept") == "application/json" || strings.HasPrefix(path, "/api/") {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing access token"})
			} else {
				c.Redirect(http.StatusTemporaryRedirect, "/auth")
			}
			c.Abort()
			return
		}
		
		c.Next()
	}
}