package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const identityKey contextKey = "identity"

// Identity holds the resolved request identity injected into context by Auth middleware.
type Identity struct {
	UserID   string
	Username string
	Role     string
}

// IdentityFromContext retrieves the Identity set by Auth middleware.
func IdentityFromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityKey).(Identity)
	return id, ok
}

// Auth middleware resolves request identity from network trust (192.168.0.0/16) or JWT Bearer token.
// Returns 401 if neither resolves a valid identity.
func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := resolveIdentity(r, jwtSecret)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), identityKey, identity)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole middleware allows only requests whose identity role is in the given list.
// Must be used after Auth middleware.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := IdentityFromContext(r.Context())
			if !ok || !allowed[identity.Role] {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ResolveIdentity resolves identity from network trust or JWT.
// Exported so auth_handler can use it for /api/auth/me without going through middleware.
func ResolveIdentity(r *http.Request, jwtSecret string) (Identity, bool) {
	return resolveIdentity(r, jwtSecret)
}

func resolveIdentity(r *http.Request, jwtSecret string) (Identity, bool) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return Identity{}, false
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return Identity{}, false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return Identity{}, false
	}
	sub, _ := claims["sub"].(string)
	username, _ := claims["username"].(string)
	role, _ := claims["role"].(string)
	if sub == "" || role == "" {
		return Identity{}, false
	}
	return Identity{UserID: sub, Username: username, Role: role}, true
}
