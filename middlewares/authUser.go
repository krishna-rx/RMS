package middlewares

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"net/http"
	"rms/database"
	"time"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "sessionToken should be provided", http.StatusUnauthorized)
			return
		}
		var expiryAt sql.NullTime
		err := database.DB.QueryRow(` SELECT archived_at FROM user_session WHERE id = $1`, authHeader).Scan(&expiryAt)
		if err != nil {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		if expiryAt.Valid && expiryAt.Time.Before(time.Now()) {
			http.Error(w, "Session expired", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "userSession_id", authHeader)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func CheckOwnership(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}
		var userId uuid.UUID
		err := database.DB.QueryRow(` SELECT user_id FROM user_session WHERE id = $1`, authHeader).Scan(&userId)
		rows, err := database.DB.Query(`
            SELECT r.role
            FROM user_roles AS ur
            JOIN roles AS r ON ur.roles_id = r.id
            WHERE ur.user_id = $1`, userId)
		if err != nil {
			http.Error(w, "cant get the user role details", http.StatusNotFound)
		}
		defer rows.Close()
		var roles []string
		for rows.Next() {
			var role string
			if err := rows.Scan(&role); err != nil {
				http.Error(w, "cant fetch the user role", http.StatusNotFound)
			}
			roles = append(roles, role)
		}
		if len(roles) == 0 {
			http.Error(w, "user has no roles", http.StatusNotFound)
		}
		ctx := context.WithValue(r.Context(), "userID", userId)
		ctx = context.WithValue(ctx, "roles", roles)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
