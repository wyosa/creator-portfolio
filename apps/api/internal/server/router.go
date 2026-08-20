package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"api/internal/auth"
	"api/internal/config"
	"api/internal/handlers"
	"api/internal/store"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config, st *store.Store) *gin.Engine {
	r := gin.Default()
	r.MaxMultipartMemory = 32 << 20
	// Trust no proxies: ClientIP must reflect RemoteAddr, not a
	// client-supplied X-Forwarded-For header (the login rate limit keys on it).
	if err := r.SetTrustedProxies(nil); err != nil {
		slog.Warn("failed to disable trusted proxies", "error", err)
	}

	// Security headers for every route, including /media static files.
	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	})

	authH := handlers.NewAuthHandler(cfg)
	mediaH := handlers.NewMediaHandler(st, cfg.MediaDir())
	uploadH := handlers.NewUploadHandler(cfg.MediaDir())
	settingsH := handlers.NewSettingsHandler(st)

	r.Static("/media", cfg.MediaDir())

	api := r.Group("/api")
	// Cap API request bodies at 1 MiB; /api/upload enforces its own 1 GiB
	// limit. FullPath is the matched route, so path tricks cannot bypass it.
	api.Use(func(c *gin.Context) {
		if c.FullPath() != "/api/upload" {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		}
		c.Next()
	})
	{
		api.GET("/health", func(c *gin.Context) {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			defer cancel()
			if err := st.Ping(ctx); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		api.POST("/auth/login", handlers.NewRateLimiter(10, time.Minute), authH.Login)
		api.POST("/auth/logout", authH.Logout)

		api.GET("/media", mediaH.List)
		api.GET("/settings", settingsH.Get)

		authed := api.Group("")
		authed.Use(authMiddleware(cfg.JWTSecret))
		{
			authed.GET("/auth/me", authH.Me)
			authed.POST("/media", mediaH.Create)
			authed.PUT("/media/reorder", mediaH.Reorder)
			authed.PUT("/media/:id", mediaH.Update)
			authed.DELETE("/media/:id", mediaH.Delete)
			authed.POST("/upload", uploadH.Upload)
			authed.PUT("/settings", settingsH.Update)
		}
	}

	return r
}

func authMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(handlers.SessionCookieName)
		if err != nil || cookie == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		username, err := auth.ParseToken(secret, cookie)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("username", username)
		c.Next()
	}
}
