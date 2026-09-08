package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"url-shortener/internal/config"
	"url-shortener/internal/db"
	"url-shortener/internal/link"
	"url-shortener/internal/server"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer database.Close()

	store, err := link.NewStore(database, 10000)
	if err != nil {
		log.Fatalf("failed to initialize link store: %v", err)
	}
	defer store.Close()

	srv := server.New(store)

	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      srv,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Server running on port %s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly")
}
