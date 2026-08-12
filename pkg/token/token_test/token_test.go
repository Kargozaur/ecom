package token_test

import (
	"crypto/rand"
	"crypto/rsa"
	"pkg/token"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func setupKeys(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v\n", err)
	}
	return privateKey, &privateKey.PublicKey
}

func testConfig(t *testing.T) *token.TokenConfig {
	t.Helper()
	accessPriv, accessPub := setupKeys(t)
	refreshPriv, refreshPub := setupKeys(t)
	return &token.TokenConfig{
		AccessTokenPrivateKey:  accessPriv,
		AccessTokenPublicKey:   accessPub,
		RefreshTokenPrivateKey: refreshPriv,
		RefreshTokenPublicKey:  refreshPub,
		AccessTokenTTL:         15 * time.Minute,
		RefreshTokenTTL:        7 * 24 * time.Hour,
		Issuer:                 "some-issuer",
	}
}

func TestGenerateAccessToken_Success(t *testing.T) {
	cfg := testConfig(t)
	gen := token.NewTokenGenerator(cfg)
	token, err := gen.GenerateAccessToken("user-123", cfg.Issuer)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	if token == "" {
		t.Fatal("empty token was generated")
	}
}

func TestGenerateRefreshToken_Success(t *testing.T) {
	cfg := testConfig(t)
	gen := token.NewTokenGenerator(cfg)
	token, err := gen.GenerateRefreshToken("user-123", cfg.Issuer)
	if err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	if token == "" {
		t.Fatal("empty token was generated")
	}
}
func TestGeneratedToken_UsesRS256(t *testing.T) {
	cfg := testConfig(t)
	gen := token.NewTokenGenerator(cfg)

	tokenStr, err := gen.GenerateAccessToken("user-123", cfg.Issuer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, _, err := jwt.NewParser().ParseUnverified(tokenStr, &token.JWTClaims{})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	if parsed.Method.Alg() != "RS256" {
		t.Fatalf("expected alg RS256, got %s", parsed.Method.Alg())
	}
}
func TestGeneratedToken_ContainsCorrectClaims(t *testing.T) {
	cfg := testConfig(t)
	gen := token.NewTokenGenerator(cfg)

	before := time.Now().UTC()
	tokenStr, err := gen.GenerateAccessToken("user-123", cfg.Issuer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, _, err := jwt.NewParser().ParseUnverified(tokenStr, &token.JWTClaims{})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	claims := parsed.Claims.(*token.JWTClaims)

	if claims.UserID != "user-123" {
		t.Errorf("expected UserID user-123, got %s", claims.UserID)
	}
	if claims.Issuer != cfg.Issuer {
		t.Errorf("expected issuer %s, got %s", cfg.Issuer, claims.Issuer)
	}
	if claims.ExpiresAt == nil || claims.ExpiresAt.Before(before.Add(cfg.AccessTokenTTL-time.Second)) {
		t.Errorf("unexpected expiry: %v", claims.ExpiresAt)
	}
}
func TestValidateToken_ValidAccessToken(t *testing.T) {
	cfg := testConfig(t)
	gen := token.NewTokenGenerator(cfg)
	val := token.NewTokenValidator(cfg)

	tokenStr, err := gen.GenerateAccessToken("user-123", cfg.Issuer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !val.ValidateToken(tokenStr, token.Access) {
		t.Fatal("expected token to be valid")
	}
}

func TestValidateToken_ValidRefreshToken(t *testing.T) {
	cfg := testConfig(t)
	gen := token.NewTokenGenerator(cfg)
	val := token.NewTokenValidator(cfg)

	tokenStr, err := gen.GenerateRefreshToken("user-123", cfg.Issuer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !val.ValidateToken(tokenStr, token.Refresh) {
		t.Fatal("expected token to be valid")
	}
}

func TestValidateToken_WrongTokenType(t *testing.T) {
	cfg := testConfig(t)
	gen := token.NewTokenGenerator(cfg)
	val := token.NewTokenValidator(cfg)

	tokenStr, err := gen.GenerateAccessToken("user-123", cfg.Issuer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val.ValidateToken(tokenStr, token.Refresh) {
		t.Fatal("expected validation to fail: access token checked against refresh key")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	cfg := testConfig(t)
	cfg.AccessTokenTTL = -1 * time.Minute
	gen := token.NewTokenGenerator(cfg)
	val := token.NewTokenValidator(cfg)

	tokenStr, err := gen.GenerateAccessToken("user-123", cfg.Issuer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val.ValidateToken(tokenStr, token.Access) {
		t.Fatal("expected expired token to be invalid")
	}
}
func TestValidateToken_SignedWithWrongKey(t *testing.T) {
	cfg := testConfig(t)
	val := token.NewTokenValidator(cfg)

	otherCfg := testConfig(t)
	otherGen := token.NewTokenGenerator(otherCfg)

	tokenStr, err := otherGen.GenerateAccessToken("user-123", cfg.Issuer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val.ValidateToken(tokenStr, token.Access) {
		t.Fatal("expected token signed with foreign key to be invalid")
	}
}
func TestValidateToken_MalformedToken(t *testing.T) {
	cfg := testConfig(t)
	val := token.NewTokenValidator(cfg)

	if val.ValidateToken("not.a.token", token.Access) {
		t.Fatal("expected token to be invalid")
	}
}
func TestValidateToken_RejectsAlgConfusion(t *testing.T) {
	cfg := testConfig(t)
	val := token.NewTokenValidator(cfg)

	claims := token.JWTClaims{
		UserID: "someID",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.Issuer,
		},
	}

	pubKeyBytes := []byte(cfg.AccessTokenPublicKey.N.String())
	hsToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := hsToken.SignedString(pubKeyBytes)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	if val.ValidateToken(tokenStr, token.Access) {
		t.Fatal("was expecting rejection")
	}
}
func TestGetUserID_Success(t *testing.T) {
	cfg := testConfig(t)
	gen := token.NewTokenGenerator(cfg)
	val := token.NewTokenValidator(cfg)

	uID := "user-123"
	tokenStr, err := gen.GenerateAccessToken(uID, cfg.Issuer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	userID, err := val.GetUserID(tokenStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != uID {
		t.Errorf("expected user-123, got %s", userID)
	}
}
func TestGetUserID_ExpiredToken(t *testing.T) {
	cfg := testConfig(t)
	cfg.AccessTokenTTL = -1 * time.Minute
	gen := token.NewTokenGenerator(cfg)
	val := token.NewTokenValidator(cfg)

	tokenStr, err := gen.GenerateAccessToken("user-123", cfg.Issuer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = val.GetUserID(tokenStr)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}
