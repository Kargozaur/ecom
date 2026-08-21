package token

import (
	"crypto/rsa"
	"encoding/base64"
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

func decodeBase64Key(value string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}

	return decoded, nil
}

func NewTokenConfig() (*TokenConfig, error) {
	accessPrivB64 := envreader.Read("ACCESS_TOKEN_PRIVATE_KEY", "")
	accessPubB64 := envreader.Read("ACCESS_TOKEN_PUBLIC_KEY", "")
	refreshPrivB64 := envreader.Read("REFRESH_TOKEN_PRIVATE_KEY", "")
	refreshPubB64 := envreader.Read("REFRESH_TOKEN_PUBLIC_KEY", "")

	if accessPrivB64 == "" ||
		accessPubB64 == "" ||
		refreshPrivB64 == "" ||
		refreshPubB64 == "" {
		return nil, errors.New("RSA keys must not be empty")
	}

	accessPrivPEM, err := decodeBase64Key(accessPrivB64)
	if err != nil {
		return nil, fmt.Errorf("decode access private key: %w", err)
	}

	accessPubPEM, err := decodeBase64Key(accessPubB64)
	if err != nil {
		return nil, fmt.Errorf("decode access public key: %w", err)
	}

	refreshPrivPEM, err := decodeBase64Key(refreshPrivB64)
	if err != nil {
		return nil, fmt.Errorf("decode refresh private key: %w", err)
	}

	refreshPubPEM, err := decodeBase64Key(refreshPubB64)
	if err != nil {
		return nil, fmt.Errorf("decode refresh public key: %w", err)
	}

	accessPriv, err := jwt.ParseRSAPrivateKeyFromPEM(accessPrivPEM)
	if err != nil {
		return nil, fmt.Errorf("parse access private key: %w", err)
	}

	accessPub, err := jwt.ParseRSAPublicKeyFromPEM(accessPubPEM)
	if err != nil {
		return nil, fmt.Errorf("parse access public key: %w", err)
	}

	refreshPriv, err := jwt.ParseRSAPrivateKeyFromPEM(refreshPrivPEM)
	if err != nil {
		return nil, fmt.Errorf("parse refresh private key: %w", err)
	}

	refreshPub, err := jwt.ParseRSAPublicKeyFromPEM(refreshPubPEM)
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
