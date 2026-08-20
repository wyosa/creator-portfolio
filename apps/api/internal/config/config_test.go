package config

import (
	"strings"
	"testing"
)

// clearEnv empties every variable Load reads, so env() falls back to defaults.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"PORT", "DATA_DIR", "ADMIN_USER", "ADMIN_PASSWORD",
		"JWT_SECRET", "COOKIE_SECURE", "GIN_MODE",
	} {
		t.Setenv(k, "")
	}
}

func TestLoadDevDefaults(t *testing.T) {
	clearEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.DataDir != "./data" {
		t.Errorf("DataDir = %q, want ./data", cfg.DataDir)
	}
	if cfg.AdminUser != "admin" {
		t.Errorf("AdminUser = %q, want admin", cfg.AdminUser)
	}
	if cfg.AdminPassword != defaultAdminPassword {
		t.Errorf("AdminPassword = %q, want the dev default", cfg.AdminPassword)
	}
	if cfg.JWTSecret != defaultJWTSecret {
		t.Errorf("JWTSecret = %q, want the dev default", cfg.JWTSecret)
	}
	if cfg.CookieSecure {
		t.Error("CookieSecure = true, want false by default")
	}
}

func TestLoadRelease(t *testing.T) {
	releaseEnv := func(t *testing.T) {
		clearEnv(t)
		t.Setenv("GIN_MODE", "release")
		t.Setenv("ADMIN_PASSWORD", "a-strong-password")
		t.Setenv("JWT_SECRET", strings.Repeat("x", minJWTSecretLen))
	}

	t.Run("valid", func(t *testing.T) {
		releaseEnv(t)
		cfg, err := Load()
		if err != nil {
			t.Fatalf("load: %v", err)
		}
		if cfg.AdminPassword != "a-strong-password" {
			t.Errorf("AdminPassword = %q", cfg.AdminPassword)
		}
	})

	t.Run("missing secrets", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("GIN_MODE", "release")
		if _, err := Load(); err == nil {
			t.Fatal("expected error for default secrets in release mode")
		}
	})

	t.Run("default admin password", func(t *testing.T) {
		releaseEnv(t)
		t.Setenv("ADMIN_PASSWORD", defaultAdminPassword)
		if _, err := Load(); err == nil {
			t.Fatal("expected error for default ADMIN_PASSWORD in release mode")
		}
	})

	t.Run("short jwt secret", func(t *testing.T) {
		releaseEnv(t)
		t.Setenv("JWT_SECRET", strings.Repeat("x", minJWTSecretLen-1))
		if _, err := Load(); err == nil {
			t.Fatal("expected error for short JWT_SECRET in release mode")
		}
	})

	t.Run("short secret accepted in dev", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("JWT_SECRET", "short")
		if _, err := Load(); err != nil {
			t.Fatalf("dev mode must accept a short secret: %v", err)
		}
	})
}

func TestLoadPortValidation(t *testing.T) {
	cases := []struct {
		name    string
		port    string
		wantErr bool
	}{
		{"default", "", false},
		{"low", "1", false},
		{"high", "65535", false},
		{"typical", "9090", false},
		{"zero", "0", true},
		{"too high", "65536", true},
		{"negative", "-1", true},
		{"not a number", "abc", true},
		{"with spaces", "80 80", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			clearEnv(t)
			t.Setenv("PORT", tc.port)
			_, err := Load()
			if (err != nil) != tc.wantErr {
				t.Fatalf("PORT=%q: err = %v, wantErr %v", tc.port, err, tc.wantErr)
			}
		})
	}
}
