package models

import (
	"database/sql/driver"
	"strings"
	"time"

	"github.com/lib/pq"
)

// StringArray is a custom type for PostgreSQL text arrays that works with Swagger
type StringArray []string

// Value implements driver.Valuer interface
func (s StringArray) Value() (driver.Value, error) {
	return pq.StringArray(s).Value()
}

// Scan implements sql.Scanner interface
func (s *StringArray) Scan(value interface{}) error {
	var arr pq.StringArray
	err := arr.Scan(value)
	if err != nil {
		return err
	}
	*s = StringArray(arr)
	return nil
}

// String returns string representation
func (s StringArray) String() string {
	return strings.Join(s, ",")
}

// Post represents a blog post
type Post struct {
	ID        uint        `json:"id" gorm:"primaryKey" example:"1"`
	Title     string      `json:"title" gorm:"not null" example:"My First Blog Post"`
	Content   string      `json:"content" gorm:"type:text;not null" example:"This is the content of my first blog post."`
	Tags      StringArray `json:"tags" gorm:"type:text[]" swaggertype:"array,string" example:"golang,programming,tutorial"`
	CreatedAt time.Time   `json:"created_at" example:"2023-09-14T08:04:38.522445Z"`
	UpdatedAt time.Time   `json:"updated_at" example:"2023-09-14T08:04:38.522445Z"`
}

// ActivityLog represents system activity logs
type ActivityLog struct {
	ID       uint      `json:"id" gorm:"primaryKey" example:"1"`
	Action   string    `json:"action" gorm:"not null" example:"new_post"`
	PostID   *uint     `json:"post_id" example:"1"` // Changed to pointer to allow NULL values
	Post     Post      `json:"post" gorm:"foreignKey:PostID"`
	LoggedAt time.Time `json:"logged_at" example:"2023-09-14T08:04:38.522445Z"`
}

// PostSearchResult represents the structure for Elasticsearch documents
type PostSearchResult struct {
	ID      uint     `json:"id" example:"1"`
	Title   string   `json:"title" example:"My First Blog Post"`
	Content string   `json:"content" example:"This is the content of my first blog post."`
	Tags    []string `json:"tags" example:"golang,programming,tutorial"`
}

// CreatePostRequest represents the request body for creating a post
type CreatePostRequest struct {
	Title   string   `json:"title" binding:"required" example:"My First Blog Post"`
	Content string   `json:"content" binding:"required" example:"This is the content of my first blog post."`
	Tags    []string `json:"tags" example:"golang,programming,tutorial"`
}

// UpdatePostRequest represents the request body for updating a post
type UpdatePostRequest struct {
	Title   string   `json:"title" example:"Updated Blog Post Title"`
	Content string   `json:"content" example:"Updated content of the blog post."`
	Tags    []string `json:"tags" example:"golang,programming,updated"`
}

// PostWithRelated represents a post with related posts
type PostWithRelated struct {
	Post         Post   `json:"post"`
	RelatedPosts []Post `json:"related_posts"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	CurrentPage int  `json:"current_page" example:"1"`
	TotalPages  int  `json:"total_pages" example:"5"`
	TotalCount  int  `json:"total_count" example:"50"`
	Limit       int  `json:"limit" example:"10"`
	HasNext     bool `json:"has_next" example:"true"`
	HasPrev     bool `json:"has_prev" example:"false"`
}

// PostsResponse represents the response for getting posts with pagination
type PostsResponse struct {
	Posts      []Post             `json:"posts"`
	Pagination PaginationResponse `json:"pagination"`
}

// ActivityLogsResponse represents the response for getting activity logs with pagination
type ActivityLogsResponse struct {
	Logs       []ActivityLog      `json:"logs"`
	Pagination PaginationResponse `json:"pagination"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid input"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

// SearchResponse represents the response for search operations
type SearchResponse struct {
	Posts []PostSearchResult `json:"posts"`
	Total int64              `json:"total" example:"25"`
	Took  int                `json:"took" example:"5"`
}

// TagSearchResponse represents the response for tag-based search
type TagSearchResponse struct {
	Posts []Post `json:"posts"`
	Count int    `json:"count" example:"10"`
}
