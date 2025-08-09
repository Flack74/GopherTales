package handlers

import (
	"html/template"
	"log"
	"net/http"
)

// SelectionHandler handles the gopher selection page requests
type SelectionHandler struct {
	templateDir string
}

// NewSelectionHandler creates a new selection handler
func NewSelectionHandler(templateDir string) *SelectionHandler {
	return &SelectionHandler{
		templateDir: templateDir,
	}
}

// ServeHTTP handles HTTP requests for the gopher selection page
func (h *SelectionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the selection template
	tmplPath := h.templateDir + "/selection.html"
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf("Error parsing selection template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute the template
	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("Error executing selection template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
