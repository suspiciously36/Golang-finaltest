package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"github.com/susbuntu/blog-api/models"
)

// CreatePost handles POST /posts - Creates a new post with transaction support
// @Summary Create a new blog post
// @Description Create a new blog post with transaction support for data integrity
// @Tags posts
// @Accept json
// @Produce json
// @Param post body models.CreatePostRequest true "Post creation request"
// @Success 201 {object} models.Post
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /posts [post]
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
		Tags:    models.StringArray(req.Tags),
	}

	if err := tx.Create(&post).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	// Create activity log
	activityLog := models.ActivityLog{
		Action: "new_post",
		PostID: &post.ID,
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
// @Summary Get a specific blog post
// @Description Retrieves a post by ID with Redis caching (5-minute TTL)
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} models.Post
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /posts/{id} [get]
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

// GetPostWithRelated handles GET /posts/:id/related - Gets a post with related posts
// @Summary Get a post with related posts
// @Description Retrieves a post by ID along with related posts based on tag similarity using Elasticsearch
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} models.PostWithRelated
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /posts/{id}/related [get]
func (h *Handler) GetPostWithRelated(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	// Get the main post from database
	var post models.Post
	if err := h.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Find related posts using Elasticsearch
	relatedPosts, err := h.findRelatedPosts(post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find related posts"})
		return
	}

	result := models.PostWithRelated{
		Post:         post,
		RelatedPosts: relatedPosts,
	}

	c.JSON(http.StatusOK, result)
}

// GetActivityLogs handles GET /activity-logs - Gets all activity logs with pagination
// @Summary Get activity logs
// @Description Retrieves all system activity logs with pagination support
// @Tags activity-logs
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} models.ActivityLogsResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /activity-logs [get]
func (h *Handler) GetActivityLogs(c *gin.Context) {
	// Parse pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var logs []models.ActivityLog
	var total int64
	
	// Get total count
	if err := h.DB.Model(&models.ActivityLog{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count activity logs"})
		return
	}

	// Get logs with pagination, ordered by logged_at descending
	if err := h.DB.Preload("Post").Order("logged_at DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch activity logs"})
		return
	}

	// Calculate pagination info
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, gin.H{
			"logs": logs,
			"pagination": gin.H{
				"current_page": page,
				"total_pages":  totalPages,
				"total_count":  total,
				"limit":        limit,
				"has_next":     hasNext,
				"has_prev":     hasPrev,
			},
		})
}

// findRelatedPosts finds posts related to the given post based on tags using Elasticsearch
func (h *Handler) findRelatedPosts(post models.Post) ([]models.Post, error) {
	ctx := context.Background()

	// If the post has no tags, return empty slice
	if len(post.Tags) == 0 {
		return []models.Post{}, nil
	}

	// Create a bool query with should clauses for each tag
	boolQuery := elastic.NewBoolQuery()
	
	// Add should clauses for each tag (OR logic)
	for _, tag := range post.Tags {
		termQuery := elastic.NewTermQuery("tags", tag)
		boolQuery = boolQuery.Should(termQuery)
	}
	
	// Exclude the current post from results
	boolQuery = boolQuery.MustNot(elastic.NewTermQuery("id", post.ID))
	
	// Set minimum should match to ensure at least one tag matches
	boolQuery = boolQuery.MinimumShouldMatch("1")

	// Execute the search
	searchResult, err := h.ES.Search().
		Index("posts").
		Query(boolQuery).
		Size(5). // Limit to 5 related posts
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("elasticsearch search failed: %v", err)
	}

	// Extract post IDs from search results
	var postIDs []uint
	for _, hit := range searchResult.Hits.Hits {
		var searchPost models.PostSearchResult
		if err := json.Unmarshal(hit.Source, &searchPost); err == nil {
			postIDs = append(postIDs, searchPost.ID)
		}
	}

	// If no related posts found, return empty slice
	if len(postIDs) == 0 {
		return []models.Post{}, nil
	}

	// Fetch full post data from database
	var relatedPosts []models.Post
	if err := h.DB.Where("id IN ?", postIDs).Find(&relatedPosts).Error; err != nil {
		return nil, fmt.Errorf("database query failed: %v", err)
	}

	return relatedPosts, nil
}

// UpdatePost handles PUT /posts/:id - Updates a post with cache invalidation
// @Summary Update a blog post
// @Description Updates a post and invalidates the cache
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Param post body models.UpdatePostRequest true "Post update request"
// @Success 200 {object} models.Post
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /posts/{id} [put]
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
		post.Tags = models.StringArray(req.Tags)
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
// @Summary Search posts by tag
// @Description Searches posts containing a specific tag using optimized GIN indexing
// @Tags posts
// @Accept json
// @Produce json
// @Param tag query string true "Tag name to search for"
// @Success 200 {object} models.TagSearchResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /posts/search-by-tag [get]
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
// @Summary Full-text search posts
// @Description Performs full-text search across post titles and content using Elasticsearch
// @Tags posts
// @Accept json
// @Produce json
// @Param q query string true "Search query string"
// @Success 200 {object} models.SearchResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /posts/search [get]
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

// GetAllPosts handles GET /posts - Gets all posts with pagination
// @Summary Get all blog posts
// @Description Retrieves all posts with pagination support
// @Tags posts
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} models.PostsResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /posts [get]
func (h *Handler) GetAllPosts(c *gin.Context) {
	// Parse pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	var posts []models.Post
	var total int64
	
	// Get total count
	if err := h.DB.Model(&models.Post{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count posts"})
		return
	}

	// Get posts with pagination, ordered by created_at descending
	if err := h.DB.Order("created_at DESC").Offset(offset).Limit(limit).Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
		return
	}

	// Calculate pagination info
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages
	hasPrev := page > 1

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
		"pagination": gin.H{
			"current_page": page,
			"total_pages":  totalPages,
			"total_count":  total,
			"limit":        limit,
			"has_next":     hasNext,
			"has_prev":     hasPrev,
		},
	})
}

// DeletePost handles DELETE /posts/:id - Deletes a post with cache invalidation
// @Summary Delete a blog post
// @Description Deletes a post and cleans up related data with cache invalidation
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /posts/{id} [delete]
func (h *Handler) DeletePost(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	// Start transaction
	tx := h.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Check if post exists
	var post models.Post
	if err := tx.First(&post, id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Delete related activity logs first
	if err := tx.Where("post_id = ?", id).Delete(&models.ActivityLog{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete activity logs"})
		return
	}

	// Delete the post
	if err := tx.Delete(&post).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}

	// Create deletion activity log AFTER deleting the post (with null PostID since post is gone)
	if err := tx.Exec("INSERT INTO activity_logs (action, post_id) VALUES ($1, NULL)", "delete_post").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create activity log"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Invalidate cache
	ctx := context.Background()
	cacheKey := fmt.Sprintf("post:%d", id)
	h.Redis.Del(ctx, cacheKey)

	// Delete from Elasticsearch
	go h.deletePostFromES(uint(id))

	c.JSON(http.StatusOK, gin.H{
		"message": "Post deleted successfully",
		"id":      id,
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

// deletePostFromES deletes a post from Elasticsearch
func (h *Handler) deletePostFromES(postID uint) {
	ctx := context.Background()

	_, err := h.ES.Delete().
		Index("posts").
		Id(fmt.Sprintf("%d", postID)).
		Do(ctx)

	if err != nil {
		fmt.Printf("Failed to delete post from Elasticsearch: %v\n", err)
	}
}
