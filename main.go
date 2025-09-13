package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/susbuntu/blog-api/config"
	"github.com/susbuntu/blog-api/database"
	"github.com/susbuntu/blog-api/routes"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database connections
	db := database.InitPostgreSQL(cfg)
	redis := database.InitRedis(cfg)
	es := database.InitElasticsearch(cfg)

	// Auto migrate database
	database.AutoMigrate(db)

	// Initialize Gin router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router, db, redis, es)

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(router.Run(":" + cfg.Port))
}
