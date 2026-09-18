package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	BaseURL     string
	Host        string
	Port        string
	DatabaseURL string
	CacheSize   int
}

func (c *Config) Address() string {
	host := c.Host
	if host == "" {
		host = "0.0.0.0"
	}
	return fmt.Sprintf("%s:%s", host, c.Port)
}

func (c *Config) URL() string {
	url := c.BaseURL
	switch url {
	case "", "0.0.0.0":
		url = "localhost"
	}
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	return fmt.Sprintf("http://%s:%s", url, c.Port)
}

func Load() *Config {
	baseURL := os.Getenv("BASE_URL")

	port := os.Getenv("PORT")
	if port == "" {
		port = "3011"
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
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
		BaseURL:     baseURL,
		Host:        host,
		Port:        port,
		DatabaseURL: dbURL,
		CacheSize:   cacheSize,
	}
}
