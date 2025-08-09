package handlers

import (
	"html/template"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"GopherTales/internal/services"
)

type DashboardHandler struct {
	userService *services.UserService
	templateDir string
}

func NewDashboardHandler(userService *services.UserService, templateDir string) *DashboardHandler {
	return &DashboardHandler{
		userService: userService,
		templateDir: templateDir,
	}
}

func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from cookie
	cookie, err := r.Cookie("user_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, err := primitive.ObjectIDFromHex(cookie.Value)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"User": user,
	}

	tmpl, err := template.ParseFiles(h.templateDir + "/dashboard.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}
