package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"api/internal/auth"
	"api/internal/config"
	"api/internal/handlers"
	"api/internal/store"

	"github.com/gin-gonic/gin"
)

const testJWTSecret = "0123456789abcdef0123456789abcdef"

func newTestRouter(t *testing.T) (*gin.Engine, *store.Store, config.Config) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	cfg := config.Config{
		Port:          "8080",
		DataDir:       dir,
		AdminUser:     "admin",
		AdminPassword: "correct-password",
		JWTSecret:     testJWTSecret,
	}
	if err := os.MkdirAll(cfg.MediaDir(), 0o750); err != nil {
		t.Fatalf("mkdir media: %v", err)
	}
	st, err := store.Open(cfg.DBPath())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return NewRouter(cfg, st), st, cfg
}

func serve(r *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHealth(t *testing.T) {
	r, st, _ := newTestRouter(t)

	w := serve(r, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("health status = %d, want 200: %s", w.Code, w.Body.String())
	}

	// A broken database must surface as 503, infra depends on it.
	if err := st.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}
	w = serve(r, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("health with closed db status = %d, want 503", w.Code)
	}
}

func TestSecurityHeaders(t *testing.T) {
	r, _, cfg := newTestRouter(t)
	if err := os.WriteFile(filepath.Join(cfg.MediaDir(), "file.txt"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write media file: %v", err)
	}

	want := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}
	for _, path := range []string{"/api/health", "/media/file.txt"} {
		w := serve(r, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want 200", path, w.Code)
		}
		for k, v := range want {
			if got := w.Header().Get(k); got != v {
				t.Errorf("GET %s: header %s = %q, want %q", path, k, got, v)
			}
		}
	}
}

func TestAuthMiddleware(t *testing.T) {
	r, _, _ := newTestRouter(t)

	// No cookie at all.
	w := serve(r, httptest.NewRequest(http.MethodGet, "/api/auth/me", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("no cookie: status = %d, want 401", w.Code)
	}

	// Garbage cookie.
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	// #nosec G124 -- client-side request cookie in a test, attributes don't apply.
	req.AddCookie(&http.Cookie{Name: handlers.SessionCookieName, Value: "garbage"})
	w = serve(r, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("garbage cookie: status = %d, want 401", w.Code)
	}

	// Valid session token.
	token, err := auth.CreateToken(testJWTSecret, "admin")
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	// #nosec G124 -- client-side request cookie in a test, attributes don't apply.
	req.AddCookie(&http.Cookie{Name: handlers.SessionCookieName, Value: token})
	w = serve(r, req)
	if w.Code != http.StatusOK {
		t.Fatalf("valid cookie: status = %d, want 200: %s", w.Code, w.Body.String())
	}
	var body struct {
		User struct {
			Username string `json:"username"`
		} `json:"user"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.User.Username != "admin" {
		t.Errorf("username = %q, want admin", body.User.Username)
	}
}

func TestLoginFlowAndRateLimit(t *testing.T) {
	r, _, _ := newTestRouter(t)

	login := func(username, password string) *httptest.ResponseRecorder {
		payload, err := json.Marshal(map[string]string{"username": username, "password": password})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		return serve(r, req)
	}

	// 10 wrong attempts pass through to the handler (401), the 11th is
	// rejected by the limiter before reaching it.
	for i := 1; i <= 10; i++ {
		if w := login("admin", "wrong"); w.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: status = %d, want 401", i, w.Code)
		}
	}
	if w := login("admin", "correct-password"); w.Code != http.StatusTooManyRequests {
		t.Fatalf("over-limit status = %d, want 429: %s", w.Code, w.Body.String())
	}
}

func TestForwardedForNotTrusted(t *testing.T) {
	r, _, _ := newTestRouter(t)

	login := func(xff string) *httptest.ResponseRecorder {
		payload, err := json.Marshal(map[string]string{"username": "admin", "password": "wrong"})
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", xff)
		return serve(r, req)
	}

	// With no trusted proxies the limiter keys on RemoteAddr, so attempts
	// with different spoofed X-Forwarded-For values share a single budget:
	// 5 + 5 wrong attempts pass, the next one is over the limit of 10.
	for i := 0; i < 5; i++ {
		if w := login("1.1.1.1"); w.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d (1.1.1.1): status = %d, want 401", i+1, w.Code)
		}
		if w := login("2.2.2.2"); w.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d (2.2.2.2): status = %d, want 401", i+1, w.Code)
		}
	}
	if w := login("3.3.3.3"); w.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429: spoofed XFF must not reset the budget", w.Code)
	}
}
