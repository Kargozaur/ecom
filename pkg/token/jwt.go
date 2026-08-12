package token

import (
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenGenerator struct {
	config *TokenConfig
}

type TokenValidator struct {
	config *TokenConfig
}

type ITokenGenerator interface {
	GenerateAccessToken(string, string) (string, error)
	GenerateRefreshToken(string, string) (string, error)
}

type ITokenValidator interface {
	ValidateToken(string, TokenType) bool
	GetUserID(string) (string, error)
}
type JWTClaims struct {
	UserID string
	jwt.RegisteredClaims
}

func NewTokenGenerator(cfg *TokenConfig) *TokenGenerator {
	return &TokenGenerator{
		config: cfg,
	}
}

func NewTokenValidator(cfg *TokenConfig) *TokenValidator {
	return &TokenValidator{
		config: cfg,
	}
}
func (c *TokenGenerator) generateToken(key *rsa.PrivateKey, userID, iss string, exp time.Duration) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(exp)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
			Issuer:    iss,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(key)
}

func (c *TokenGenerator) GenerateAccessToken(userID, iss string) (string, error) {
	token, err := c.generateToken(c.config.AccessTokenPrivateKey, userID, iss, c.config.AccessTokenTTL)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (c *TokenGenerator) GenerateRefreshToken(userID, iss string) (string, error) {
	token, err := c.generateToken(c.config.RefreshTokenPrivateKey, userID, iss, c.config.RefreshTokenTTL)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (v *TokenValidator) parseToken(jwtToken string, tokenType TokenType) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &JWTClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, ErrInvalidMethod
		}
		switch tokenType {
		case Access:
			return v.config.AccessTokenPublicKey, nil
		case Refresh:
			return v.config.RefreshTokenPublicKey, nil
		}
		return nil, ErrTokenInvalid
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (v *TokenValidator) ValidateToken(jwtToken string, tokenType TokenType) bool {
	token, err := v.parseToken(jwtToken, tokenType)
	if err != nil || !token.Valid {
		return false
	}
	return true
}

func (v *TokenValidator) GetUserID(jwtToken string) (string, error) {
	token, err := v.parseToken(jwtToken, Access)
	if err != nil || !token.Valid {
		return "", err
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return "", ErrInvalidClaims
	}
	return claims.UserID, nil
}
