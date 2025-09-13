package config

import (
	"os"
)

type Config struct {
	Port     string
	Database DatabaseConfig
	Redis    RedisConfig
	ES       ElasticsearchConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type RedisConfig struct {
	Host string
	Port string
}

type ElasticsearchConfig struct {
	Host string
	Port string
}

func Load() *Config {
	return &Config{
		Port: getEnv("PORT", "8080"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "blog_user"),
			Password: getEnv("DB_PASSWORD", "blog_password"),
			Name:     getEnv("DB_NAME", "blog_db"),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"),
			Port: getEnv("REDIS_PORT", "6379"),
		},
		ES: ElasticsearchConfig{
			Host: getEnv("ES_HOST", "localhost"),
			Port: getEnv("ES_PORT", "9200"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
