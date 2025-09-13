package handlers

import (
	"github.com/go-redis/redis/v8"
	"github.com/olivere/elastic/v7"
	"gorm.io/gorm"
)

type Handler struct {
	DB    *gorm.DB
	Redis *redis.Client
	ES    *elastic.Client
}

func NewHandler(db *gorm.DB, redis *redis.Client, es *elastic.Client) *Handler {
	return &Handler{
		DB:    db,
		Redis: redis,
		ES:    es,
	}
}
