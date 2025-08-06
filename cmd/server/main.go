package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"GopherTales/internal/config"
	"GopherTales/internal/handlers"
	"GopherTales/internal/middleware"
	"GopherTales/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize services
	storyService := services.NewStoryService(cfg.Story.DataFile)

	// Load story data
	if err := storyService.LoadStory(); err != nil {
		log.Fatalf("Failed to load story: %v", err)
	}

	log.Printf("Successfully loaded story with %d arcs", len(storyService.GetAvailableArcs()))

	// Initialize handlers
	homeHandler := handlers.NewHomeHandler(cfg.Story.TemplateDir)
	storyHandler := handlers.NewStoryHandler(storyService, cfg.Story.TemplateDir)
	apiHandler := handlers.NewAPIHandler(storyService)

	// Setup routes
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir(cfg.Story.StaticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Web routes
	mux.Handle("/", homeHandler)
	mux.Handle("/story", storyHandler)

	// API routes
	mux.HandleFunc("/api/health", apiHandler.HealthCheck)
	mux.HandleFunc("/api/stats", apiHandler.GetStoryStats)
	mux.HandleFunc("/api/arcs", apiHandler.GetAllArcs)
	mux.HandleFunc("/api/arc", apiHandler.GetArc)

	// Apply middleware
	handler := middleware.Chain(
		mux,
		middleware.Logger,
		middleware.Recovery,
		middleware.SecurityHeaders,
		middleware.CORS,
	)

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting GopherTales server on %s", cfg.Address())
		log.Printf("Visit http://%s to start your adventure!", cfg.Address())

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server gracefully stopped")
	}
}
