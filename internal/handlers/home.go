package handlers

import (
	"html/template"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"GopherTales/internal/models"
	"GopherTales/internal/services"
)

// HomeHandler handles the home page requests
type HomeHandler struct {
	templateDir string
	userService *services.UserService
}

// NewHomeHandler creates a new home handler
func NewHomeHandler(templateDir string, userService *services.UserService) *HomeHandler {
	return &HomeHandler{
		templateDir: templateDir,
		userService: userService,
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

	// Check if user is logged in
	var user *models.User
	if cookie, err := r.Cookie("user_id"); err == nil {
		if userID, err := primitive.ObjectIDFromHex(cookie.Value); err == nil {
			if u, err := h.userService.GetUserByID(userID); err == nil {
				user = u
			}
		}
	}

	data := map[string]interface{}{
		"User":       user,
		"IsLoggedIn": user != nil,
	}

	// Execute the template
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing home template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
