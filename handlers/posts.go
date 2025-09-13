package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/olivere/elastic/v7"
	"github.com/susbuntu/blog-api/models"
)

// CreatePost handles POST /posts - Creates a new post with transaction support
func (h *Handler) CreatePost(c *gin.Context) {
	var req models.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start transaction
	tx := h.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Create post
	post := models.Post{
		Title:   req.Title,
		Content: req.Content,
		Tags:    pq.StringArray(req.Tags),
	}

	if err := tx.Create(&post).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	// Create activity log
	activityLog := models.ActivityLog{
		Action: "new_post",
		PostID: post.ID,
	}

	if err := tx.Create(&activityLog).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create activity log"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Index in Elasticsearch
	go h.indexPostInES(post)

	c.JSON(http.StatusCreated, post)
}

// GetPost handles GET /posts/:id - Gets a post with Cache-Aside pattern
func (h *Handler) GetPost(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("post:%d", id)

	// Try to get from Redis first (Cache-Aside pattern)
	cachedData, err := h.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache hit - return cached data
		var post models.Post
		if json.Unmarshal([]byte(cachedData), &post) == nil {
			c.JSON(http.StatusOK, post)
			return
		}
	}

	// Cache miss - get from database
	var post models.Post
	if err := h.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Cache the result with 5 minutes TTL
	postJSON, _ := json.Marshal(post)
	h.Redis.Set(ctx, cacheKey, postJSON, 5*time.Minute)

	c.JSON(http.StatusOK, post)
}

// UpdatePost handles PUT /posts/:id - Updates a post with cache invalidation
func (h *Handler) UpdatePost(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var req models.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find existing post
	var post models.Post
	if err := h.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Update fields if provided
	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Content != "" {
		post.Content = req.Content
	}
	if req.Tags != nil {
		post.Tags = pq.StringArray(req.Tags)
	}

	// Save to database
	if err := h.DB.Save(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}

	// Invalidate cache
	ctx := context.Background()
	cacheKey := fmt.Sprintf("post:%d", id)
	h.Redis.Del(ctx, cacheKey)

	// Update in Elasticsearch
	go h.indexPostInES(post)

	c.JSON(http.StatusOK, post)
}

// SearchPostsByTag handles GET /posts/search-by-tag?tag=<tag_name>
func (h *Handler) SearchPostsByTag(c *gin.Context) {
	tag := c.Query("tag")
	if tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag parameter is required"})
		return
	}

	var posts []models.Post
	// Use GIN index for efficient tag searching
	err := h.DB.Where("tags @> ARRAY[?]", tag).Find(&posts).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search posts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
		"count": len(posts),
	})
}

// SearchPosts handles GET /posts/search?q=<query_string>
func (h *Handler) SearchPosts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q parameter is required"})
		return
	}

	ctx := context.Background()

	// Create multi-match query for title and content
	searchQuery := elastic.NewMultiMatchQuery(query, "title", "content").
		Type("best_fields").
		Fuzziness("AUTO")

	searchResult, err := h.ES.Search().
		Index("posts").
		Query(searchQuery).
		Size(50).
		Do(ctx)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	var posts []models.PostSearchResult
	for _, hit := range searchResult.Hits.Hits {
		var post models.PostSearchResult
		if err := json.Unmarshal(hit.Source, &post); err == nil {
			posts = append(posts, post)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
		"total": searchResult.Hits.TotalHits.Value,
		"took":  searchResult.TookInMillis,
	})
}

// indexPostInES indexes a post in Elasticsearch
func (h *Handler) indexPostInES(post models.Post) {
	ctx := context.Background()

	doc := models.PostSearchResult{
		ID:      post.ID,
		Title:   post.Title,
		Content: post.Content,
		Tags:    []string(post.Tags),
	}

	_, err := h.ES.Index().
		Index("posts").
		Id(fmt.Sprintf("%d", post.ID)).
		BodyJson(doc).
		Do(ctx)

	if err != nil {
		fmt.Printf("Failed to index post in Elasticsearch: %v\n", err)
	}
}
