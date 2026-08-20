package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api/internal/auth"
	"api/internal/config"

	"github.com/gin-gonic/gin"
)

func newAuthTestRouter(maxAttempts int) *gin.Engine {
	gin.SetMode(gin.TestMode)
	cfg := config.Config{
		AdminUser:     "admin",
		AdminPassword: "correct-password",
		JWTSecret:     "0123456789abcdef0123456789abcdef",
	}
	h := NewAuthHandler(cfg)
	r := gin.New()
	r.POST("/login", NewRateLimiter(maxAttempts, time.Minute), h.Login)
	return r
}

func sessionCookie(t *testing.T, w *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, c := range w.Result().Cookies() {
		if c.Name == SessionCookieName {
			return c
		}
	}
	t.Fatal("no session cookie in response")
	return nil
}

func TestLoginSuccess(t *testing.T) {
	r := newAuthTestRouter(10)
	w := doJSON(t, r, http.MethodPost, "/login", gin.H{
		"username": "admin", "password": "correct-password",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}

	cookie := sessionCookie(t, w)
	if !cookie.HttpOnly {
		t.Error("session cookie must be HttpOnly")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, want Lax", cookie.SameSite)
	}
	username, err := auth.ParseToken("0123456789abcdef0123456789abcdef", cookie.Value)
	if err != nil {
		t.Fatalf("session token does not parse: %v", err)
	}
	if username != "admin" {
		t.Errorf("token subject = %q, want admin", username)
	}

	var body struct {
		User struct {
			Username string `json:"username"`
		} `json:"user"`
	}
	decodeBody(t, w, &body)
	if body.User.Username != "admin" {
		t.Errorf("body user = %q, want admin", body.User.Username)
	}
}

func TestLoginRejects(t *testing.T) {
	r := newAuthTestRouter(10)
	cases := []struct {
		name string
		body any
	}{
		{"wrong password", gin.H{"username": "admin", "password": "wrong"}},
		{"wrong username", gin.H{"username": "root", "password": "correct-password"}},
		{"empty fields", gin.H{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := doJSON(t, r, http.MethodPost, "/login", tc.body)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401: %s", w.Code, w.Body.String())
			}
			for _, c := range w.Result().Cookies() {
				if c.Name == SessionCookieName && c.Value != "" {
					t.Error("session cookie must not be set on failed login")
				}
			}
		})
	}
}

func TestLoginRateLimit(t *testing.T) {
	const max = 3
	r := newAuthTestRouter(max)

	for i := 1; i <= max; i++ {
		w := doJSON(t, r, http.MethodPost, "/login", gin.H{"username": "admin", "password": "wrong"})
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: status = %d, want 401", i, w.Code)
		}
	}
	w := doJSON(t, r, http.MethodPost, "/login", gin.H{"username": "admin", "password": "correct-password"})
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("over-limit status = %d, want 429: %s", w.Code, w.Body.String())
	}
	var body struct {
		Error string `json:"error"`
	}
	decodeBody(t, w, &body)
	if body.Error != "too many attempts" {
		t.Errorf("error = %q, want %q", body.Error, "too many attempts")
	}
}
