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

	"pivotal/internal/config"
	"pivotal/internal/db"
	"pivotal/internal/job"
	"pivotal/internal/link"
	"pivotal/internal/server"
	"pivotal/web"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	linkRepo := link.NewRepository(database, nil, nil)
	linkSvc, err := link.NewService(linkRepo, cfg.CacheSize, cfg.URL())
	if err != nil {
		log.Fatalf("failed to initialize link service: %v", err)
	}
	defer linkSvc.Close()

	jobRepository := job.NewRepository(database)
	jobExecutor := job.NewJobExecutor(linkSvc, jobRepository)
	jobManager := job.NewManager(jobRepository, jobExecutor, 3)
	jobManager.Start()

	defer jobManager.Stop()

	webFS, err := web.Assets()
	if err != nil {
		log.Fatalf("failed to load web assets: %v", err)
	}

	srv := server.NewServer(linkSvc, webFS, jobManager)

	httpServer := &http.Server{
		Addr:         cfg.Address(),
		Handler:      srv,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("server starting on port %s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server error: %v", err)
		}
	}()

	<-shutdown
	log.Println("server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("error shutting down http server: %v", err)
	}

	log.Println("server stopped cleanly")
}
