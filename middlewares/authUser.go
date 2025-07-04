package middlewares

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"os"
	"rms/database"
	"rms/database/dbHelper"
	"rms/models"
	"strings"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			http.Error(w, "invalid token format dont found bearer", http.StatusUnauthorized)
			return
		}
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		rawUserID, exists := claims["user_id"]

		if !exists || rawUserID == nil {
			http.Error(w, "user_id claim missing", http.StatusUnauthorized)
			return
		}

		userID, ok := rawUserID.(string)
		if !ok {
			http.Error(w, "user_id claim is not a string", http.StatusUnauthorized)
			return
		}
		var rolesString []string
		err = dbHelper.GetAllRoleByID(&rolesString, userID)
		var roles []models.Role
		for _, role := range rolesString {
			roles = append(roles, models.Role(role))
		}
		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "roles", roles)
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
