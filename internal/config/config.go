package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Host        string
	Port        string
	DatabaseURL string
	CacheSize   int
}

func (c *Config) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func Load() *Config {
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}

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
		Host:        host,
		Port:        port,
		DatabaseURL: dbURL,
		CacheSize:   cacheSize,
	}
}
