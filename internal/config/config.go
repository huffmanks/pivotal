package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        string
	DatabaseURL string
	CacheSize   int
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "data.db"
	}

	cacheSize := 1000
	if val := os.Getenv("CACHE_SIZE"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			cacheSize = parsed
		}
	}

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
		CacheSize:   cacheSize,
	}
}
