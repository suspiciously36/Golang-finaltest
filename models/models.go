package models

import (
	"time"

	"github.com/lib/pq"
)

type Post struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Title     string         `json:"title" gorm:"not null"`
	Content   string         `json:"content" gorm:"type:text;not null"`
	Tags      pq.StringArray `json:"tags" gorm:"type:text[]"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type ActivityLog struct {
	ID       uint      `json:"id" gorm:"primaryKey"`
	Action   string    `json:"action" gorm:"not null"`
	PostID   uint      `json:"post_id"`
	Post     Post      `json:"post" gorm:"foreignKey:PostID"`
	LoggedAt time.Time `json:"logged_at"`
}

// PostSearchResult represents the structure for Elasticsearch documents
type PostSearchResult struct {
	ID      uint     `json:"id"`
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

// CreatePostRequest represents the request body for creating a post
type CreatePostRequest struct {
	Title   string   `json:"title" binding:"required"`
	Content string   `json:"content" binding:"required"`
	Tags    []string `json:"tags"`
}

// UpdatePostRequest represents the request body for updating a post
type UpdatePostRequest struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

// PostWithRelated represents a post with related posts
type PostWithRelated struct {
	Post         Post   `json:"post"`
	RelatedPosts []Post `json:"related_posts"`
}
