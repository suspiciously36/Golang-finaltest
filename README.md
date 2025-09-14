# Blog API - Golang Final Project

A comprehensive blog API built with Golang, featuring PostgreSQL, Redis caching, and Elasticsearch for full-text search.

## Features

- **PostgreSQL**: Optimized database with GIN indexing for tag searches and transaction support
- **Redis**: Cache-Aside pattern implementation for improved read performance
- **Elasticsearch**: Full-text search capabilities across post titles and content
- **Related Posts**: Intelligent content discovery based on tag similarity
- **Activity Logging**: Comprehensive system activity tracking with pagination
- **Swagger Documentation**: Interactive API documentation with testing capabilities
- **GORM**: Database ORM for easy data management
- **Gin**: Fast HTTP web framework

## Architecture

- **Database Layer**: PostgreSQL with optimized indexing
- **Caching Layer**: Redis with Cache-Aside pattern and TTL
- **Search Layer**: Elasticsearch for full-text search
- **API Layer**: RESTful API with Gin framework

## Requirements

- Docker and Docker Compose
- Go 1.21+ (for local development)

## Quick Start

### 1. Clone and Start Services

```bash
# Clone the repository
git clone <your-repo-url>
cd blog-api

# Start all services with Docker Compose
docker-compose up -d

# Wait for services to be ready (about 30-60 seconds)
# Check if all services are running
docker-compose ps
```

### 2. Verify Services

Check if all services are healthy:

```bash
# Check API health
curl http://localhost:8080/health

# Check Elasticsearch
curl http://localhost:9200/_cluster/health

# Check Redis (requires redis-cli)
docker exec blog_redis redis-cli ping
```

**Expected Responses:**
- API Health: `{"message":"Blog API is running","status":"OK"}`
- Elasticsearch: `{"status":"green"}` or `{"status":"yellow"}`
- Redis: `PONG`

### 3. Quick Test

Verify the API is working with a simple test:

```bash
# Create a test post
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Hello World",
    "content": "My first blog post using this API!",
    "tags": ["hello", "test", "api"]
  }'

# Get all posts
curl http://localhost:8080/api/v1/posts
```

### 4. Access Swagger Documentation

Open your browser and visit the **interactive API documentation**:

**🔗 [Swagger UI](http://localhost:8080/swagger/index.html)**

You can test all endpoints directly from the Swagger interface!

### 4. Access Swagger Documentation

The API now includes interactive Swagger documentation for easy testing:

**🚀 Swagger UI: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)**

The Swagger UI provides:
- 📋 **Complete API documentation** with request/response schemas
- 🧪 **Interactive testing** - Try all endpoints directly from the browser
- 📖 **Request examples** and parameter descriptions
- 🔍 **Schema explorer** for all data models

## API Documentation

### Swagger UI

The API includes comprehensive **Swagger/OpenAPI documentation** with an interactive UI for easy testing:

**🔗 Access Swagger UI at: `http://localhost:8080/swagger/index.html`**

**Features:**
- 📋 **Complete API Reference** - All endpoints documented with examples
- 🧪 **Interactive Testing** - Test APIs directly from the browser
- 📝 **Request/Response Schemas** - Detailed models and examples
- 🔍 **Parameter Documentation** - Query parameters, path parameters, and body schemas
- ✅ **Validation Rules** - Required fields and data types clearly marked

### API Endpoints

### 📖 Interactive Documentation: [Swagger UI](http://localhost:8080/swagger/index.html)

For the best API testing experience, use the **Swagger UI** which provides:
- Interactive endpoint testing
- Complete request/response schemas
- Real-time API exploration
- Built-in parameter validation

### Base URL: `http://localhost:8080/api/v1`

### 1. Create Post (with Transaction)

Creates a new post and logs the activity in a single transaction.

```bash
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My First Blog Post",
    "content": "This is the content of my first blog post. It contains some interesting information about technology.",
    "tags": ["technology", "golang", "api"]
  }'
```

### 2. Get Post (with Cache-Aside)

Retrieves a post by ID with Redis caching (5-minute TTL).

```bash
curl http://localhost:8080/api/v1/posts/1
```

### 3. Update Post (with Cache Invalidation)

Updates a post and invalidates the cache.

```bash
curl -X PUT http://localhost:8080/api/v1/posts/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Blog Post Title",
    "content": "Updated content with new information.",
    "tags": ["technology", "golang", "api", "updated"]
  }'
```

### 4. Search Posts by Tag (PostgreSQL GIN Index)

Searches posts containing a specific tag using optimized GIN indexing.

```bash
curl "http://localhost:8080/api/v1/posts/search-by-tag?tag=golang"
```

### 5. Full-text Search (Elasticsearch)

Performs full-text search across post titles and content.

```bash
curl "http://localhost:8080/api/v1/posts/search?q=technology"
```

### 6. Get All Posts (with Pagination)

Retrieves all posts with pagination support.

```bash
# Get first page (default: 10 posts per page)
curl "http://localhost:8080/api/v1/posts"

# Get specific page with custom limit
curl "http://localhost:8080/api/v1/posts?page=2&limit=5"
```

### 7. Related Posts (Bonus Feature)

Gets a post with related posts based on tag similarity using Elasticsearch.

```bash
curl "http://localhost:8080/api/v1/posts/1/related"
```

### 8. Activity Logs (with Pagination)

Retrieves system activity logs with pagination.

```bash
# Get recent activity logs
curl "http://localhost:8080/api/v1/activity-logs"

# Get specific page with custom limit
curl "http://localhost:8080/api/v1/activity-logs?page=1&limit=10"
```

### 9. Delete Post (with Cache Invalidation)

Deletes a post and cleans up related data.

```bash
curl -X DELETE "http://localhost:8080/api/v1/posts/1"
```

## Complete API Reference

| Method | Endpoint | Description | Required Body |
|--------|----------|-------------|---------------|
| `GET` | `/health` | Health check endpoint | - |
| `POST` | `/api/v1/posts` | Create new post | `{title, content, tags}` |
| `GET` | `/api/v1/posts` | Get all posts (paginated) | - |
| `GET` | `/api/v1/posts/:id` | Get specific post (cached) | - |
| `GET` | `/api/v1/posts/:id/related` | Get post with related posts | - |
| `PUT` | `/api/v1/posts/:id` | Update post | `{title?, content?, tags?}` |
| `DELETE` | `/api/v1/posts/:id` | Delete post | - |
| `GET` | `/api/v1/posts/search-by-tag?tag=<tag>` | Search by tag (GIN index) | - |
| `GET` | `/api/v1/posts/search?q=<query>` | Full-text search | - |
| `GET` | `/api/v1/activity-logs` | Get activity logs (paginated) | - |

### Query Parameters

**Pagination (for `/posts` and `/activity-logs`):**
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 10 for posts, 20 for logs, max: 100)

**Search:**
- `tag`: Tag name for tag-based search
- `q`: Query string for full-text search

## Testing the Implementation

### 1. Test Database Transaction

```bash
# Create a post - this should create both post and activity log
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Transaction Test",
    "content": "Testing transaction functionality",
    "tags": ["test", "transaction"]
  }'

# Verify in database (optional - requires psql)
docker exec blog_postgres psql -U blog_user -d blog_db -c "SELECT * FROM posts; SELECT * FROM activity_logs;"
```

### 2. Test Cache-Aside Pattern

```bash
# First request - cache miss (slower)
time curl http://localhost:8080/api/v1/posts/1

# Second request - cache hit (faster)
time curl http://localhost:8080/api/v1/posts/1

# Check Redis cache
docker exec blog_redis redis-cli get "post:1"
```

### 3. Test Cache Invalidation

```bash
# Get post (to cache it)
curl http://localhost:8080/api/v1/posts/1

# Verify cache exists
docker exec blog_redis redis-cli get "post:1"

# Update post (should invalidate cache)
curl -X PUT http://localhost:8080/api/v1/posts/1 \
  -H "Content-Type: application/json" \
  -d '{"title": "Cache Invalidation Test"}'

# Verify cache is cleared
docker exec blog_redis redis-cli get "post:1"
```

### 4. Test GIN Index Performance

```bash
# Create multiple posts with tags
for i in {1..5}; do
  curl -X POST http://localhost:8080/api/v1/posts \
    -H "Content-Type: application/json" \
    -d "{
      \"title\": \"Post $i\",
      \"content\": \"Content for post $i\",
      \"tags\": [\"tag$i\", \"common\", \"test\"]
    }"
done

# Search by tag (should be fast due to GIN index)
curl "http://localhost:8080/api/v1/posts/search-by-tag?tag=common"
```

### 5. Test Elasticsearch Full-text Search

```bash
# Create posts with different content
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Golang Best Practices",
    "content": "Learn about the best practices in Golang development including error handling and performance optimization.",
    "tags": ["golang", "programming", "best-practices"]
  }'

# Wait a moment for Elasticsearch indexing
sleep 2

# Search for content
curl "http://localhost:8080/api/v1/posts/search?q=golang"
curl "http://localhost:8080/api/v1/posts/search?q=best%20practices"
curl "http://localhost:8080/api/v1/posts/search?q=performance"
```

### 6. Test Related Posts Feature (Bonus)

```bash
# Create posts with overlapping tags
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Advanced Go Concurrency",
    "content": "Deep dive into Go concurrency patterns and best practices.",
    "tags": ["golang", "concurrency", "advanced"]
  }'

curl -X POST http://localhost:8080/api/v1/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Go Channels Guide",
    "content": "Complete guide to using channels in Go programming.",
    "tags": ["golang", "channels", "programming"]
  }'

# Wait for Elasticsearch indexing
sleep 3

# Test related posts (should find posts with similar tags)
curl "http://localhost:8080/api/v1/posts/1/related"
```

### 7. Test Activity Logs

```bash
# View activity logs (shows post creation, deletion activities)
curl "http://localhost:8080/api/v1/activity-logs"

# View with pagination
curl "http://localhost:8080/api/v1/activity-logs?page=1&limit=5"
```

### 8. Test Complete Workflow

```bash
# 1. Create a post
POST_ID=$(curl -s -X POST http://localhost:8080/api/v1/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Workflow Test Post",
    "content": "Testing the complete API workflow.",
    "tags": ["test", "workflow", "api"]
  }' | jq -r '.id')

echo "Created post with ID: $POST_ID"

# 2. Get the post (cache miss)
curl "http://localhost:8080/api/v1/posts/$POST_ID"

# 3. Get again (cache hit)
curl "http://localhost:8080/api/v1/posts/$POST_ID"

# 4. Update the post (invalidates cache)
curl -X PUT "http://localhost:8080/api/v1/posts/$POST_ID" \
  -H "Content-Type: application/json" \
  -d '{"title": "Updated Workflow Test"}'

# 5. Search by tag
curl "http://localhost:8080/api/v1/posts/search-by-tag?tag=workflow"

# 6. Full-text search
curl "http://localhost:8080/api/v1/posts/search?q=workflow"

# 7. Get related posts
curl "http://localhost:8080/api/v1/posts/$POST_ID/related"

# 8. View activity logs
curl "http://localhost:8080/api/v1/activity-logs?limit=3"
```

## Database Schema

### Posts Table

```sql
CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- GIN index for fast tag searches
CREATE INDEX idx_posts_tags ON posts USING GIN(tags);
```

### Activity Logs Table

```sql
CREATE TABLE activity_logs (
    id SERIAL PRIMARY KEY,
    action VARCHAR(100) NOT NULL,
    post_id INTEGER REFERENCES posts(id),
    logged_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Performance Features

1. **GIN Index**: Fast tag-based searches using PostgreSQL's GIN indexing
2. **Redis Caching**: 5-minute TTL cache for post retrieval with Cache-Aside pattern
3. **Cache Invalidation**: Automatic cache clearing on updates and deletes
4. **Transaction Safety**: ACID compliance for post creation and activity logging
5. **Elasticsearch**: Fast full-text search with fuzzy matching and related posts discovery
6. **Optimized Pagination**: Efficient pagination with metadata for all list endpoints
7. **Async Processing**: Background indexing for Elasticsearch operations

## Development

### Local Development Setup

```bash
# Install dependencies
go mod download

# Run locally (requires services to be running via Docker Compose)
go run main.go
```

### Environment Variables

- `DB_HOST`: PostgreSQL host (default: localhost)
- `DB_PORT`: PostgreSQL port (default: 5432)
- `DB_USER`: PostgreSQL user (default: blog_user)
- `DB_PASSWORD`: PostgreSQL password (default: blog_password)
- `DB_NAME`: PostgreSQL database name (default: blog_db)
- `REDIS_HOST`: Redis host (default: localhost)
- `REDIS_PORT`: Redis port (default: 6379)
- `ES_HOST`: Elasticsearch host (default: localhost)
- `ES_PORT`: Elasticsearch port (default: 9200)
- `PORT`: API server port (default: 8080)

## Project Structure

```
├── main.go                 # Application entry point
├── docker-compose.yml      # Docker services configuration
├── Dockerfile             # API service container
├── init.sql              # Database initialization
├── config/
│   └── config.go         # Configuration management
├── database/
│   └── database.go       # Database connections
├── models/
│   └── models.go         # Data models
├── handlers/
│   ├── handler.go        # Handler initialization
│   └── posts.go          # Post-related handlers
└── routes/
    └── routes.go         # Route definitions
```

## Troubleshooting

### Common Issues

1. **Services not starting**: Wait longer for Elasticsearch to initialize
2. **Connection refused**: Ensure all services are running with `docker-compose ps`
3. **Elasticsearch mapping issues**: Check Elasticsearch logs with `docker-compose logs elasticsearch`
4. **Database connection issues**: Verify PostgreSQL is ready with `docker-compose logs postgres`

### Logs

```bash
# View all service logs
docker-compose logs

# View specific service logs
docker-compose logs blog_api
docker-compose logs postgres
docker-compose logs redis
docker-compose logs elasticsearch
```

## Next Steps (Additional Enhancements)

The project is **COMPLETE** with all required features implemented, including the bonus "Related Posts" feature:

✅ **Implemented Features:**
- Elasticsearch bool queries with should clauses for related posts
- Tag-based similarity matching with exclusion of current post
- Activity logs with comprehensive pagination
- Advanced caching strategies with proper invalidation

**Potential Future Enhancements:**
- User authentication and authorization (JWT)
- Comment system for posts  
- File upload capabilities for post attachments
- Real-time notifications with WebSockets
- API rate limiting and security enhancements
- Advanced analytics and reporting

## License

This project is created for educational purposes as a final Golang project.
