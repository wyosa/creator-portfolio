package config

import (
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	Port          string
	DataDir       string
	AdminUser     string
	AdminPassword string
	JWTSecret     string
	CookieSecure  bool
}

const (
	defaultAdminPassword = "admin"
	defaultJWTSecret     = "dev-secret-change-me"
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() Config {
	cfg := Config{
		Port:          env("PORT", "8080"),
		DataDir:       env("DATA_DIR", "./data"),
		AdminUser:     env("ADMIN_USER", "admin"),
		AdminPassword: env("ADMIN_PASSWORD", defaultAdminPassword),
		JWTSecret:     env("JWT_SECRET", defaultJWTSecret),
		CookieSecure:  env("COOKIE_SECURE", "") == "true",
	}
	if os.Getenv("GIN_MODE") == "release" {
		if cfg.AdminPassword == defaultAdminPassword || cfg.JWTSecret == defaultJWTSecret {
			log.Fatal("config: ADMIN_PASSWORD and JWT_SECRET must be set to non-default values in release mode")
		}
	} else {
		if cfg.AdminPassword == defaultAdminPassword {
			log.Printf("config: warning: ADMIN_PASSWORD is the default dev value")
		}
		if cfg.JWTSecret == defaultJWTSecret {
			log.Printf("config: warning: JWT_SECRET is the default dev value")
		}
	}
	return cfg
}

func (c Config) DBPath() string {
	return filepath.Join(c.DataDir, "portfolio.db")
}

func (c Config) MediaDir() string {
	return filepath.Join(c.DataDir, "media")
}
