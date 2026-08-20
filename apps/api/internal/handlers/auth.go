package handlers

import (
	"crypto/subtle"
	"net/http"
	"time"

	"api/internal/auth"
	"api/internal/config"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	cfg config.Config
}

func NewAuthHandler(cfg config.Config) *AuthHandler {
	return &AuthHandler{cfg: cfg}
}

// SessionCookieName is the cookie carrying the admin JWT.
const SessionCookieName = "portfolio_session"

func (h *AuthHandler) setCookie(c *gin.Context, value string, maxAge int) {
	// #nosec G124 -- HttpOnly and SameSite=Lax are always set; Secure is
	// config-driven (COOKIE_SECURE) so local dev can run over plain HTTP.
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     SessionCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		Expires:  time.Now().Add(time.Duration(maxAge) * time.Second),
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "invalid request body")
		return
	}
	userOK := subtle.ConstantTimeCompare([]byte(req.Username), []byte(h.cfg.AdminUser)) == 1
	passOK := subtle.ConstantTimeCompare([]byte(req.Password), []byte(h.cfg.AdminPassword)) == 1
	if !userOK || !passOK {
		fail(c, http.StatusUnauthorized, "invalid credentials")
		return
	}
	token, err := auth.CreateToken(h.cfg.JWTSecret, req.Username)
	if err != nil {
		failErr(c, http.StatusInternalServerError, "failed to create token", err)
		return
	}
	h.setCookie(c, token, int(auth.TokenTTL.Seconds()))
	c.JSON(http.StatusOK, gin.H{"user": gin.H{"username": req.Username}})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	h.setCookie(c, "", -1)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	username := c.GetString("username")
	if username == "" {
		fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": gin.H{"username": username}})
}
