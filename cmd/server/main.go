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
	"GopherTales/internal/database"
	"GopherTales/internal/handlers"
	"GopherTales/internal/middleware"
	"GopherTales/internal/services"
)

func main() {
	// Load .env file
	if err := config.LoadEnvFile(".env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Load configuration
	cfg := config.Load()

	// Validate MongoDB URI
	if cfg.Database.MongoURI == "" {
		log.Fatalf("MONGO_URI is required in .env file")
	}

	// Initialize databases
	mongoDB, err := database.NewMongoDB(cfg.Database.MongoURI, cfg.Database.DBName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoDB.Close()
	log.Printf("✅ Connected to MongoDB: %s", cfg.Database.DBName)

	// Initialize services
	storyService := services.NewStoryService(cfg.Story.DataFile)
	userService := services.NewUserService(mongoDB)

	// Load story data
	if err := storyService.LoadStory(); err != nil {
		log.Fatalf("Failed to load story: %v", err)
	}

	log.Printf("Successfully loaded story with %d arcs", len(storyService.GetAvailableArcs()))

	// Initialize handlers
	homeHandler := handlers.NewHomeHandler(cfg.Story.TemplateDir, userService)
	selectionHandler := handlers.NewSelectionHandler(cfg.Story.TemplateDir)
	storyHandler := handlers.NewStoryHandler(storyService, userService, cfg.Story.TemplateDir)
	apiHandler := handlers.NewAPIHandler(storyService)
	authHandler := handlers.NewAuthHandler(userService)
	dashboardHandler := handlers.NewDashboardHandler(userService, cfg.Story.TemplateDir)
	profileHandler := handlers.NewProfileHandler(userService, storyService, cfg.Story.TemplateDir)

	// Auth middleware
	requireAuth := middleware.RequireAuth(userService)

	// Setup routes
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir(cfg.Story.StaticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Web routes
	mux.Handle("/", homeHandler)
	mux.Handle("/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, cfg.Story.TemplateDir+"/login.html")
	}))
	mux.Handle("/register", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, cfg.Story.TemplateDir+"/register.html")
	}))
	mux.Handle("/dashboard", requireAuth(dashboardHandler))
	mux.Handle("/selection", selectionHandler)
	mux.Handle("/story", storyHandler)
	mux.Handle("/profile", requireAuth(profileHandler))

	// API routes
	mux.HandleFunc("/api/health", apiHandler.HealthCheck)
	mux.HandleFunc("/api/stats", apiHandler.GetStoryStats)
	mux.HandleFunc("/api/arcs", apiHandler.GetAllArcs)
	mux.HandleFunc("/api/arc", apiHandler.GetArc)
	mux.HandleFunc("/api/gophers", apiHandler.GetGophers)
	mux.HandleFunc("/api/gopher-stats", apiHandler.GetGopherStats)

	// Auth routes
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.HandleFunc("/api/auth/logout", authHandler.Logout)
	mux.HandleFunc("/api/bookmark", authHandler.AddBookmark)

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
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Log server startup info
	log.Printf("🚀 Starting GopherTales server on %s", cfg.Address())
	log.Printf("🌐 Visit http://%s to start your adventure!", cfg.Address())

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
