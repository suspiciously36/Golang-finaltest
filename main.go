// @title Blog API
// @version 1.0
// @description A comprehensive blog API built with Golang, featuring PostgreSQL, Redis caching, and Elasticsearch for full-text search.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @schemes http
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/susbuntu/blog-api/config"
	"github.com/susbuntu/blog-api/database"
	"github.com/susbuntu/blog-api/routes"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/susbuntu/blog-api/docs" // This will be auto-generated
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

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	log.Printf("Swagger documentation available at: http://localhost:%s/swagger/index.html", cfg.Port)
	log.Fatal(router.Run(":" + cfg.Port))
}
