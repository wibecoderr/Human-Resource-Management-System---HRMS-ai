package middleware

import (
	"hrms/dbhelper"
	"hrms/utils"
	"net/http"
	"strings"
)

// Authenticate validates JWT and checks session is still active in DB.
func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.RespondError(w, http.StatusUnauthorized, nil, "Missing or invalid Authorization header")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := utils.ParseJWT(tokenStr)
		if err != nil {
			utils.RespondError(w, http.StatusUnauthorized, err, "Invalid token")
			return
		}

		// Check session is still active in DB (handles logout)
		valid, err := dbhelper.SessionValid(claims.SessionID)
		if err != nil || !valid {
			utils.RespondError(w, http.StatusUnauthorized, nil, "Session expired or logged out")
			return
		}

		// Pass claims via request header so handlers can read them
		r.Header.Set("X-User-ID", claims.UserID)
		r.Header.Set("X-User-Role", claims.Role)
		r.Header.Set("X-Session-ID", claims.SessionID)

		next.ServeHTTP(w, r)
	})
}

// RequireRole returns 403 if the user's role is not in the allowed list.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Header.Get("X-User-Role")
			for _, role := range roles {
				if userRole == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			utils.RespondError(w, http.StatusForbidden, nil, "Access denied: insufficient role")
		})
	}
}
