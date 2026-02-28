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

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/thirana/url-shortener/internal/httpapi"
	"github.com/thirana/url-shortener/internal/shortener"
	"github.com/thirana/url-shortener/internal/store"
	storepostgres "github.com/thirana/url-shortener/internal/store/postgres"
)

func main() {
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		log.Printf("warning: failed to load .env: %v", err)
	}

	// Production default: release mode reduces debug noise and overhead.
	// You can also set: export GIN_MODE=release
	gin.SetMode(gin.ReleaseMode)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	linkStore, closeStore, err := buildStore()
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}

	svc := shortener.NewService(linkStore)
	links := httpapi.NewLinksHandler(svc)
	router := httpapi.NewRouter(links)

	log.Printf("starting api on :%s", port)
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Fatalf("server failed: %v", err)
	case sig := <-stop:
		log.Printf("shutdown signal received: %s", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}

	if err := closeStore(); err != nil {
		log.Printf("store close failed: %v", err)
	}
}

func buildStore() (store.LinkStore, func() error, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Printf("using in-memory store (set DATABASE_URL to enable Postgres)")
		return store.NewMemoryStore(), func() error { return nil }, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pgStore, err := storepostgres.New(ctx, databaseURL)
	if err != nil {
		return nil, nil, err
	}
	if err := pgStore.Ping(ctx); err != nil {
		pgStore.Close()
		return nil, nil, err
	}

	log.Printf("using postgres store")
	return pgStore, func() error {
		pgStore.Close()
		return nil
	}, nil
}
