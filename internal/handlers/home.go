package handlers

import (
	"html/template"
	"log"
	"net/http"
)

// HomeHandler handles the home page requests
type HomeHandler struct {
	templateDir string
}

// NewHomeHandler creates a new home handler
func NewHomeHandler(templateDir string) *HomeHandler {
	return &HomeHandler{
		templateDir: templateDir,
	}
}

// ServeHTTP handles HTTP requests for the home page
func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the home template
	tmplPath := h.templateDir + "/home.html"
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf("Error parsing home template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute the template
	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("Error executing home template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
