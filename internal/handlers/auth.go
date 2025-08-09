package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"GopherTales/internal/models"
	"GopherTales/internal/services"
)

type AuthHandler struct {
	userService *services.UserService
}

func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{userService: userService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Register(req.Name, req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    user.ID.Hex(),
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"user":         user,
		"redirect_url": "/dashboard",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    user.ID.Hex(),
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"user":         user,
		"redirect_url": "/dashboard",
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func (h *AuthHandler) AddBookmark(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from cookie
	cookie, err := r.Cookie("user_id")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(cookie.Value)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		Gopher string `json:"gopher"`
		Arc    string `json:"arc"`
		Title  string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create bookmark
	bookmark := models.Bookmark{
		Gopher:    req.Gopher,
		Arc:       req.Arc,
		Title:     req.Title,
		Timestamp: time.Now(),
	}

	// Add bookmark
	if err := h.userService.AddBookmark(userID, bookmark); err != nil {
		http.Error(w, "Failed to add bookmark", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Bookmark added successfully",
	})
}
