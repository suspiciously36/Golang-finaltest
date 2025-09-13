package database

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/olivere/elastic/v7"
	"github.com/susbuntu/blog-api/config"
	"github.com/susbuntu/blog-api/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitPostgreSQL(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	log.Println("Successfully connected to PostgreSQL")
	return db
}

func InitRedis(cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		DB:   0,
	})

	// Test connection
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	log.Println("Successfully connected to Redis")
	return rdb
}

func InitElasticsearch(cfg *config.Config) *elastic.Client {
	url := fmt.Sprintf("http://%s:%s", cfg.ES.Host, cfg.ES.Port)
	
	client, err := elastic.NewClient(
		elastic.SetURL(url),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
	)
	if err != nil {
		log.Fatal("Failed to connect to Elasticsearch:", err)
	}

	// Test connection
	ctx := context.Background()
	_, _, err = client.Ping(url).Do(ctx)
	if err != nil {
		log.Fatal("Failed to ping Elasticsearch:", err)
	}

	log.Println("Successfully connected to Elasticsearch")
	
	// Create index if it doesn't exist
	createPostsIndex(client)
	
	return client
}

func AutoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(&models.Post{}, &models.ActivityLog{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database migration completed")
}

func createPostsIndex(client *elastic.Client) {
	ctx := context.Background()
	
	// Check if index exists
	exists, err := client.IndexExists("posts").Do(ctx)
	if err != nil {
		log.Printf("Error checking if index exists: %v", err)
		return
	}
	
	if !exists {
		// Create index with mapping
		mapping := `{
			"mappings": {
				"properties": {
					"id": {
						"type": "integer"
					},
					"title": {
						"type": "text",
						"analyzer": "standard"
					},
					"content": {
						"type": "text",
						"analyzer": "standard"
					},
					"tags": {
						"type": "keyword"
					}
				}
			}
		}`
		
		_, err := client.CreateIndex("posts").BodyString(mapping).Do(ctx)
		if err != nil {
			log.Printf("Error creating posts index: %v", err)
		} else {
			log.Println("Posts index created successfully")
		}
	}
}
