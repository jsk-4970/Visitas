package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/visitas/backend/pkg/auth"
)

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	// UserIDContextKey is the context key for storing user ID
	UserIDContextKey ContextKey = "user_id"
	// UserEmailContextKey is the context key for storing user email
	UserEmailContextKey ContextKey = "user_email"
	// UserClaimsContextKey is the context key for storing custom claims
	UserClaimsContextKey ContextKey = "user_claims"
)

// AuthMiddleware is a middleware that verifies Firebase ID tokens
type AuthMiddleware struct {
	firebaseClient *auth.FirebaseClient
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(firebaseClient *auth.FirebaseClient) *AuthMiddleware {
	return &AuthMiddleware{
		firebaseClient: firebaseClient,
	}
}

// RequireAuth is a middleware that requires authentication for all requests
func (am *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Expected format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		idToken := parts[1]

		// Verify the ID token
		token, err := am.firebaseClient.VerifyIDToken(r.Context(), idToken)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add user information to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDContextKey, token.UID)
		if email, ok := token.Claims["email"].(string); ok {
			ctx = context.WithValue(ctx, UserEmailContextKey, email)
		}
		ctx = context.WithValue(ctx, UserClaimsContextKey, token.Claims)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth is a middleware that attempts to authenticate but doesn't require it
func (am *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				idToken := parts[1]
				token, err := am.firebaseClient.VerifyIDToken(r.Context(), idToken)
				if err == nil {
					// Add user information to context if token is valid
					ctx := r.Context()
					ctx = context.WithValue(ctx, UserIDContextKey, token.UID)
					if email, ok := token.Claims["email"].(string); ok {
						ctx = context.WithValue(ctx, UserEmailContextKey, email)
					}
					ctx = context.WithValue(ctx, UserClaimsContextKey, token.Claims)
					r = r.WithContext(ctx)
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// RequireRole is a middleware that requires a specific role claim
func (am *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserClaimsContextKey).(map[string]interface{})
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userRole, ok := claims["role"].(string)
			if !ok || userRole != role {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(string)
	return userID, ok
}

// GetUserEmailFromContext extracts user email from context
func GetUserEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailContextKey).(string)
	return email, ok
}

// GetUserClaimsFromContext extracts custom claims from context
func GetUserClaimsFromContext(ctx context.Context) (map[string]interface{}, bool) {
	claims, ok := ctx.Value(UserClaimsContextKey).(map[string]interface{})
	return claims, ok
}
