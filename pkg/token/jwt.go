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
	GenerateAccessToken(userID, iss, email string) (string, error)
	GenerateRefreshToken(userID, iss, email string) (string, error)
}

type ITokenValidator interface {
	GetClaims(string, TokenType) (*Claims, error)
}
type JWTClaims struct {
	UserID string
	Email  string
	jwt.RegisteredClaims
}

type Claims struct {
	UserID string `json:"userID"`
	Email  string `json:"email"`
}

func NewTokenGenerator(cfg *TokenConfig, serviceName string) *TokenGenerator {
	cfg.issuer = serviceName
	return &TokenGenerator{
		config: cfg,
	}
}

func NewTokenValidator(cfg *TokenConfig) *TokenValidator {
	return &TokenValidator{
		config: cfg,
	}
}
func (c *TokenGenerator) generateToken(key *rsa.PrivateKey, userID, iss, email string, exp time.Duration) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
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

func (c *TokenGenerator) GetIssuer() string {
	return c.config.issuer
}

func (c *TokenGenerator) GenerateAccessToken(userID, iss, email string) (string, error) {
	token, err := c.generateToken(c.config.AccessTokenPrivateKey, userID, iss, email, c.config.AccessTokenTTL)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (c *TokenGenerator) GenerateRefreshToken(userID, iss, email string) (string, error) {
	token, err := c.generateToken(c.config.RefreshTokenPrivateKey, userID, iss, email, c.config.RefreshTokenTTL)
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

func (v *TokenValidator) GetClaims(jwtToken string, tokenType TokenType) (*Claims, error) {
	token, err := v.parseToken(jwtToken, tokenType)
	if err != nil || !token.Valid {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, ErrInvalidClaims
	}
	return &Claims{
		UserID: claims.UserID,
		Email:  claims.Email,
	}, nil
}
