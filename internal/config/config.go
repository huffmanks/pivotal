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
		port = "3011"
	}

	dbURL := os.Getenv("DB_PATH")
	if dbURL == "" {
		dbURL = "./data/pivotal.db"
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
