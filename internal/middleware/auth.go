package middleware

import (
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"GopherTales/internal/services"
)

func RequireAuth(userService *services.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			_, err = userService.GetUserByID(userID)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
