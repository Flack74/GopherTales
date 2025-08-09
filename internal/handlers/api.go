package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"GopherTales/internal/models"
	"GopherTales/internal/services"
)

// APIHandler handles API requests
type APIHandler struct {
	storyService *services.StoryService
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(storyService *services.StoryService) *APIHandler {
	return &APIHandler{
		storyService: storyService,
	}
}

// HealthCheck handles health check requests
func (a *APIHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]any{
		"status":  "healthy",
		"service": "GopherTales",
		"version": "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding health check response: %v", err)
	}
}

// GetStoryStats returns statistics about the loaded story
func (a *APIHandler) GetStoryStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := a.storyService.GetStoryStats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("Error encoding story stats response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// GetAllArcs returns all available story arcs
func (a *APIHandler) GetAllArcs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	storyData := a.storyService.GetStoryData()
	if storyData == nil || storyData.Arcs == nil {
		http.Error(w, "Story not loaded", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"arcs": storyData.Arcs,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding all arcs response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// GetArc returns a specific story arc
func (a *APIHandler) GetArc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	arcName := r.URL.Query().Get("name")
	gopher := r.URL.Query().Get("gopher")

	if arcName == "" {
		http.Error(w, "Arc name is required", http.StatusBadRequest)
		return
	}

	var arc models.Arc
	var finalArcName string
	var err error

	if gopher != "" {
		arc, finalArcName, err = a.storyService.GetGopherArc(gopher, arcName)
	} else {
		arc, finalArcName, err = a.storyService.GetArc(arcName)
	}

	if err != nil {
		log.Printf("Error getting arc '%s': %v", arcName, err)
		http.Error(w, "Arc not found", http.StatusNotFound)
		return
	}

	response := map[string]any{
		"arc_name": finalArcName,
		"arc":      arc,
		"gopher":   gopher,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding arc response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// GetGophers returns all available gopher characters
func (a *APIHandler) GetGophers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gophers := a.storyService.GetAvailableGophers()

	response := map[string]any{
		"gophers": gophers,
		"count":   len(gophers),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding gophers response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// GetGopherStats returns statistics for all gopher stories
func (a *APIHandler) GetGopherStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := a.storyService.GetGopherStats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("Error encoding gopher stats response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
