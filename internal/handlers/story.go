package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"GopherTales/internal/models"
	"GopherTales/internal/services"
)

// StoryHandler handles story-related HTTP requests
type StoryHandler struct {
	storyService *services.StoryService
	userService  *services.UserService
	templateDir  string
}

// NewStoryHandler creates a new story handler
func NewStoryHandler(storyService *services.StoryService, userService *services.UserService, templateDir string) *StoryHandler {
	return &StoryHandler{
		storyService: storyService,
		userService:  userService,
		templateDir:  templateDir,
	}
}

// ServeHTTP handles HTTP requests for story pages
func (h *StoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get parameters from query
	arcName := r.URL.Query().Get("arc")
	gopher := r.URL.Query().Get("gopher")

	var arc models.Arc
	var finalArcName string
	var err error

	// Handle gopher-based stories
	if gopher != "" {
		arc, finalArcName, err = h.storyService.GetGopherArc(gopher, arcName)
		if err != nil {
			log.Printf("Error getting gopher arc '%s' for gopher '%s': %v", arcName, gopher, err)
			http.Error(w, "Story not available", http.StatusNotFound)
			return
		}
	} else {
		// Fallback to original story structure
		if arcName != "" && !h.storyService.ValidateArc(arcName) {
			http.Error(w, "Story arc not found", http.StatusNotFound)
			return
		}

		arc, finalArcName, err = h.storyService.GetArc(arcName)
		if err != nil {
			log.Printf("Error getting arc '%s': %v", arcName, err)
			http.Error(w, "Story not available", http.StatusInternalServerError)
			return
		}
	}

	// Check if client wants JSON response
	if r.Header.Get("Accept") == "application/json" || r.URL.Query().Get("format") == "json" {
		h.serveJSON(w, arc, finalArcName, gopher)
		return
	}

	// Serve HTML response
	h.serveHTML(w, r, arc, finalArcName, gopher)
}

// serveHTML renders the story as HTML
func (h *StoryHandler) serveHTML(w http.ResponseWriter, r *http.Request, arc models.Arc, arcName, gopher string) {
	// Parse the story template
	tmplPath := h.templateDir + "/story.html"
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf("Error parsing story template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get user for progress tracking
	var user *models.User
	if cookie, err := r.Cookie("user_id"); err == nil {
		if userID, err := primitive.ObjectIDFromHex(cookie.Value); err == nil {
			if u, err := h.userService.GetUserByID(userID); err == nil {
				user = u
				// Update progress if gopher story
				if gopher != "" {
					// Simple progress calculation based on arc depth
					progressValue := 10 // Base progress per arc
					if len(arc.Options) == 0 {
						progressValue = 100 // Ending arc
					}
					h.userService.UpdateProgress(userID, gopher, progressValue)
				}
			}
		}
	}

	// Create page data
	pageData := models.PageData{
		Arc:     arc,
		ArcName: arcName,
		Gopher:  gopher,
		User:    user,
	}

	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute the template
	if err := tmpl.Execute(w, pageData); err != nil {
		log.Printf("Error executing story template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// serveJSON returns the story data as JSON
func (h *StoryHandler) serveJSON(w http.ResponseWriter, arc models.Arc, arcName, gopher string) {
	response := map[string]any{
		"arc_name": arcName,
		"arc":      arc,
		"gopher":   gopher,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
