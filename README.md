# Blog API - Golang Final Project

A comprehensive blog API built with Golang, featuring PostgreSQL, Redis caching, and Elasticsearch for full-text search.

## Features

- **PostgreSQL**: Optimized database with GIN indexing for tag searches and transaction support
- **Redis**: Cache-Aside pattern implementation for improved read performance
- **Elasticsearch**: Full-text search capabilities across post titles and content
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

## API Endpoints

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
2. **Redis Caching**: 5-minute TTL cache for post retrieval
3. **Cache Invalidation**: Automatic cache clearing on updates
4. **Transaction Safety**: ACID compliance for post creation and logging
5. **Elasticsearch**: Fast full-text search with fuzzy matching

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

## Next Steps (Bonus Features)

The project is ready for implementing the bonus "Related Posts" feature:

- Elasticsearch bool queries with should clauses
- Tag-based similarity matching
- Exclusion of the current post from results

## License

This project is created for educational purposes as a final Golang project.
