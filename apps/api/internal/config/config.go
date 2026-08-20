package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
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
	// minJWTSecretLen keeps brute-forcing the HMAC key impractical.
	minJWTSecretLen = 32
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() (Config, error) {
	cfg := Config{
		Port:          env("PORT", "8080"),
		DataDir:       env("DATA_DIR", "./data"),
		AdminUser:     env("ADMIN_USER", "admin"),
		AdminPassword: env("ADMIN_PASSWORD", defaultAdminPassword),
		JWTSecret:     env("JWT_SECRET", defaultJWTSecret),
		CookieSecure:  env("COOKIE_SECURE", "") == "true",
	}
	port, err := strconv.Atoi(cfg.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("config: PORT %q is not a valid port number (1-65535)", cfg.Port)
	}
	if os.Getenv("GIN_MODE") == "release" {
		if cfg.AdminPassword == defaultAdminPassword {
			return Config{}, errors.New("config: ADMIN_PASSWORD must be set to a non-default value in release mode")
		}
		if cfg.JWTSecret == defaultJWTSecret || len(cfg.JWTSecret) < minJWTSecretLen {
			return Config{}, fmt.Errorf("config: JWT_SECRET must be at least %d characters in release mode", minJWTSecretLen)
		}
		if !cfg.CookieSecure {
			// Not fatal: local release runs (task local) serve plain HTTP.
			slog.Warn("config: COOKIE_SECURE is not true in release mode, session cookies will go over plain HTTP")
		}
	} else {
		if cfg.AdminPassword == defaultAdminPassword {
			slog.Warn("config: ADMIN_PASSWORD is the default dev value")
		}
		if cfg.JWTSecret == defaultJWTSecret {
			slog.Warn("config: JWT_SECRET is the default dev value")
		}
	}
	return cfg, nil
}

func (c Config) DBPath() string {
	return filepath.Join(c.DataDir, "portfolio.db")
}

func (c Config) MediaDir() string {
	return filepath.Join(c.DataDir, "media")
}
