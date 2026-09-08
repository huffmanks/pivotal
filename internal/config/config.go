package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Port   string
	DBPath string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3011"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		if userDir, err := os.UserConfigDir(); err == nil {
			dbPath = filepath.Join(userDir, "url-shortener", "shortener.db")
		} else {
			dbPath = "./data/shortener.db"
		}
	}

	return Config{
		Port:   port,
		DBPath: dbPath,
	}
}
