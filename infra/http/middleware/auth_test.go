package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mw "DartScheduler/infra/http/middleware"

	"github.com/golang-jwt/jwt/v5"
)

func makeJWT(secret, userID, username, role string, exp time.Time) string {
	claims := jwt.MapClaims{
		"sub":      userID,
		"username": username,
		"role":     role,
		"exp":      exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))
	return signed
}

func TestResolveIdentity_LocalNetwork(t *testing.T) {
	tests := []struct {
		addr    string
		isLocal bool
	}{
		{"192.168.1.100:54321", true},
		{"192.168.0.1:80", true},
		{"10.0.0.1:8080", false},
		{"127.0.0.1:8080", false},
		{"8.8.8.8:443", false},
	}
	for _, tc := range tests {
		r := &http.Request{RemoteAddr: tc.addr, Header: http.Header{}}
		id, ok := mw.ResolveIdentity(r, "secret")
		if tc.isLocal {
			if !ok || id.Role != "admin" || id.Username != "lokaal netwerk" {
				t.Errorf("addr=%q: expected local-admin identity, got ok=%v id=%+v", tc.addr, ok, id)
			}
		} else {
			if ok {
				t.Errorf("addr=%q: expected ok=false for non-local without JWT", tc.addr)
			}
		}
	}
}

func TestResolveIdentity_ValidJWT(t *testing.T) {
	secret := "test-secret"
	token := makeJWT(secret, "user-1", "alice", "admin", time.Now().Add(time.Hour))
	r := &http.Request{
		RemoteAddr: "10.0.0.1:1234",
		Header:     http.Header{"Authorization": {"Bearer " + token}},
	}
	id, ok := mw.ResolveIdentity(r, secret)
	if !ok {
		t.Fatal("expected ok=true for valid JWT")
	}
	if id.Username != "alice" || id.Role != "admin" || id.UserID != "user-1" {
		t.Errorf("unexpected identity: %+v", id)
	}
}

func TestResolveIdentity_ExpiredJWT(t *testing.T) {
	secret := "test-secret"
	token := makeJWT(secret, "user-1", "alice", "admin", time.Now().Add(-time.Hour))
	r := &http.Request{
		RemoteAddr: "10.0.0.1:1234",
		Header:     http.Header{"Authorization": {"Bearer " + token}},
	}
	_, ok := mw.ResolveIdentity(r, secret)
	if ok {
		t.Error("expected ok=false for expired token")
	}
}

func TestResolveIdentity_WrongSecret(t *testing.T) {
	token := makeJWT("secret-a", "user-1", "alice", "admin", time.Now().Add(time.Hour))
	r := &http.Request{
		RemoteAddr: "10.0.0.1:1234",
		Header:     http.Header{"Authorization": {"Bearer " + token}},
	}
	_, ok := mw.ResolveIdentity(r, "secret-b")
	if ok {
		t.Error("expected ok=false for wrong secret")
	}
}

func TestAuth_Middleware_LocalNetworkGetsAdmin(t *testing.T) {
	handler := mw.Auth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := mw.IdentityFromContext(r.Context())
		if !ok {
			http.Error(w, "no identity", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(id.Role))
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.5:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
	if rr.Body.String() != "admin" {
		t.Errorf("role: got %q, want %q", rr.Body.String(), "admin")
	}
}

func TestAuth_Middleware_NoAuth_Returns401(t *testing.T) {
	handler := mw.Auth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}

func TestRequireRole_BlocksWrongRole(t *testing.T) {
	secret := "secret"
	token := makeJWT(secret, "u1", "bob", "viewer", time.Now().Add(time.Hour))
	handler := mw.Auth(secret)(mw.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403", rr.Code)
	}
}

func TestRequireRole_AllowsCorrectRole(t *testing.T) {
	secret := "secret"
	token := makeJWT(secret, "u1", "bob", "maintainer", time.Now().Add(time.Hour))
	handler := mw.Auth(secret)(mw.RequireRole("maintainer", "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}
