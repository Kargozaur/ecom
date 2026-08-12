package token

import (
	"errors"
)

var (
	ErrTokenInvalid  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("token has expired")
	ErrInvalidClaims = errors.New("invalid token claims")
	ErrInvalidMethod = errors.New("invalid signing method")
)
