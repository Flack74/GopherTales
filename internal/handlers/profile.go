package handlers

import (
	"html/template"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"GopherTales/internal/services"
)

type ProfileHandler struct {
	userService  *services.UserService
	storyService *services.StoryService
	templateDir  string
}

func NewProfileHandler(userService *services.UserService, storyService *services.StoryService, templateDir string) *ProfileHandler {
	return &ProfileHandler{
		userService:  userService,
		storyService: storyService,
		templateDir:  templateDir,
	}
}

func (h *ProfileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	// Get all stats
	storyStats := h.storyService.GetStoryStats()
	gopherStats := h.storyService.GetGopherStats()
	issues := h.storyService.ValidateStoryIntegrity()

	// Calculate user's total progress
	totalProgress := 0
	progressCount := 0
	for _, progress := range user.Progress {
		totalProgress += progress
		progressCount++
	}
	avgProgress := 0
	if progressCount > 0 {
		avgProgress = totalProgress / progressCount
	}

	data := map[string]interface{}{
		"User":          user,
		"StoryStats":    storyStats,
		"GopherStats":   gopherStats,
		"Issues":        issues,
		"HasIssues":     len(issues) > 0,
		"TotalProgress": avgProgress,
		"BookmarkCount": len(user.Bookmarks),
	}

	tmpl, err := template.ParseFiles(h.templateDir + "/profile.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}
