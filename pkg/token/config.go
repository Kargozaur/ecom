package token

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"pkg/envreader"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenConfig struct {
	AccessTokenPrivateKey  *rsa.PrivateKey
	RefreshTokenPrivateKey *rsa.PrivateKey
	AccessTokenPublicKey   *rsa.PublicKey
	RefreshTokenPublicKey  *rsa.PublicKey
	AccessTokenTTL         time.Duration
	RefreshTokenTTL        time.Duration
	issuer                 string
}

func NewTokenConfig() (*TokenConfig, error) {
	accessPrivPEM := envreader.Read("ACCESS_TOKEN_PRIVATE_KEY", "")
	accessPubPEM := envreader.Read("ACCESS_TOKEN_PUBLIC_KEY", "")
	refreshPrivPEM := envreader.Read("REFRESH_TOKEN_PRIVATE_KEY", "")
	refreshPubPEM := envreader.Read("REFRESH_TOKEN_PUBLIC_KEY", "")

	if accessPrivPEM == "" || accessPubPEM == "" || refreshPrivPEM == "" || refreshPubPEM == "" {
		return nil, errors.New("RSA keys must not be empty")
	}
	accessPriv, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(accessPrivPEM))
	if err != nil {
		return nil, fmt.Errorf("parse access private key: %w", err)
	}
	accessPub, err := jwt.ParseRSAPublicKeyFromPEM([]byte(accessPubPEM))
	if err != nil {
		return nil, fmt.Errorf("parse access public key: %w", err)
	}
	refreshPriv, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(refreshPrivPEM))
	if err != nil {
		return nil, fmt.Errorf("parse refresh private key: %w", err)
	}
	refreshPub, err := jwt.ParseRSAPublicKeyFromPEM([]byte(refreshPubPEM))
	if err != nil {
		return nil, fmt.Errorf("parse refresh public key: %w", err)
	}

	return &TokenConfig{
		AccessTokenPrivateKey:  accessPriv,
		AccessTokenPublicKey:   accessPub,
		RefreshTokenPrivateKey: refreshPriv,
		RefreshTokenPublicKey:  refreshPub,
		AccessTokenTTL:         15 * time.Minute,
		RefreshTokenTTL:        7 * 24 * time.Hour,
	}, nil
}
