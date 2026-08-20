package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func signedToken(t *testing.T, secret string, claims jwt.RegisteredClaims) string {
	t.Helper()
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return tok
}

func TestCreateParseRoundtrip(t *testing.T) {
	tok, err := CreateToken(testSecret, "admin")
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	subject, err := ParseToken(testSecret, tok)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if subject != "admin" {
		t.Fatalf("subject = %q, want %q", subject, "admin")
	}
}

func TestParseTokenRejects(t *testing.T) {
	valid, err := CreateToken(testSecret, "admin")
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	expired := signedToken(t, testSecret, jwt.RegisteredClaims{
		Subject:   "admin",
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
	})

	noExpiry := signedToken(t, testSecret, jwt.RegisteredClaims{Subject: "admin"})

	otherSecret, err := CreateToken("another-secret-that-is-long-enough", "admin")
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	noneTok, err := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{
		Subject:   "admin",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign alg=none token: %v", err)
	}

	// Same header and claims as the valid token, but a garbage signature.
	parts := strings.Split(valid, ".")
	forged := parts[0] + "." + parts[1] + "." + "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

	// Valid signature, empty subject.
	noSubject := signedToken(t, testSecret, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})

	cases := []struct {
		name  string
		token string
	}{
		{"expired", expired},
		{"missing expiry", noExpiry},
		{"wrong secret", otherSecret},
		{"alg none", noneTok},
		{"forged signature", forged},
		{"empty subject", noSubject},
		{"garbage", "not-a-token"},
		{"empty", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ParseToken(testSecret, tc.token); err == nil {
				t.Fatal("ParseToken: expected error, got nil")
			}
		})
	}
}
