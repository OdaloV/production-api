package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"production-api/internal/config"
	"production-api/internal/handlers"
	"production-api/internal/middleware"
)

func main() {
	// load configuration
	cfg := config.Load()

	// setup logger
	slog.Info("starting server",
		"port", cfg.Port,
		"rate_limit", cfg.RateLimit,
		"rate_window_seconds", cfg.RateWindowSeconds,
		"timeout_seconds", cfg.TimeoutSeconds,
		"allowed_origins", cfg.AllowedOrigins,
	)

	// create chi router
	r := chi.NewRouter()

	// apply global middleware in order
	r.Use(middleware.Recovery)                                                                   // first - catches panics
	r.Use(middleware.RequestID)                                                                  // adds request id
	r.Use(middleware.RealIP)                                                                     // extracts real ip
	r.Use(middleware.Logger)                                                                     // logs requests
	r.Use(middleware.Timeout(time.Duration(cfg.TimeoutSeconds) * time.Second))                   // request timeout
	r.Use(middleware.RateLimit(cfg.RateLimit, time.Duration(cfg.RateWindowSeconds)*time.Second)) // rate limiting
	r.Use(middleware.CORS(cfg.AllowedOrigins))                                                   // cors headers
	r.Use(middleware.Compress)                                                                   // last - compresses response

	// register routes
	r.Get("/health", handlers.HealthCheck)
	r.Get("/users", handlers.GetUsers)
	r.Post("/users", handlers.CreateUser)
	r.Get("/users/{id}", handlers.GetUserByID)

	// create http server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// start server in goroutine
	go func() {
		slog.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// create channel to listen for OS signals
	quit := make(chan os.Signal, 1)
	// listen for SIGINT (Ctrl+C) and SIGTERM (kill)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	slog.Info("signal received, starting shutdown", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	slog.Info("server exited , all connections closed")
}
