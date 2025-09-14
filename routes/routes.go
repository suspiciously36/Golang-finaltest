package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/olivere/elastic/v7"
	"github.com/susbuntu/blog-api/handlers"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, redis *redis.Client, es *elastic.Client) {
	// Initialize handler
	h := handlers.NewHandler(db, redis, es)

	// API routes group
	api := router.Group("/api/v1")
	{
		// Posts routes
		posts := api.Group("/posts")
		{
			posts.POST("", h.CreatePost)
			posts.GET("", h.GetAllPosts)
			posts.GET("/:id", h.GetPost)
			posts.GET("/:id/related", h.GetPostWithRelated)
			posts.PUT("/:id", h.UpdatePost)
			posts.DELETE("/:id", h.DeletePost)
			posts.GET("/search-by-tag", h.SearchPostsByTag)
			posts.GET("/search", h.SearchPosts)
		}

		// Activity logs routes
		api.GET("/activity-logs", h.GetActivityLogs)
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "OK",
			"message": "Blog API is running",
		})
	})
}
