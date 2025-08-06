package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"GopherTales/internal/models"
	"GopherTales/internal/services"
)

// StoryHandler handles story-related HTTP requests
type StoryHandler struct {
	storyService *services.StoryService
	templateDir  string
}

// NewStoryHandler creates a new story handler
func NewStoryHandler(storyService *services.StoryService, templateDir string) *StoryHandler {
	return &StoryHandler{
		storyService: storyService,
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

	// Get arc name from query parameter
	arcName := r.URL.Query().Get("arc")

	// Validate arc name
	if arcName != "" && !h.storyService.ValidateArc(arcName) {
		http.Error(w, "Story arc not found", http.StatusNotFound)
		return
	}

	// Get the arc data
	arc, finalArcName, err := h.storyService.GetArc(arcName)
	if err != nil {
		log.Printf("Error getting arc '%s': %v", arcName, err)
		http.Error(w, "Story not available", http.StatusInternalServerError)
		return
	}

	// Check if client wants JSON response
	if r.Header.Get("Accept") == "application/json" || r.URL.Query().Get("format") == "json" {
		h.serveJSON(w, arc, finalArcName)
		return
	}

	// Serve HTML response
	h.serveHTML(w, arc, finalArcName)
}

// serveHTML renders the story as HTML
func (h *StoryHandler) serveHTML(w http.ResponseWriter, arc models.Arc, arcName string) {
	// Parse the story template
	tmplPath := h.templateDir + "/story.html"
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf("Error parsing story template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Create page data
	pageData := models.PageData{
		Arc:     arc,
		ArcName: arcName,
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
func (h *StoryHandler) serveJSON(w http.ResponseWriter, arc models.Arc, arcName string) {
	response := map[string]any{
		"arc_name": arcName,
		"arc":      arc,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
