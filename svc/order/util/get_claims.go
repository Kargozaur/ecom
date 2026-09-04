package util

import (
	"context"
	"order/interceptor"
	"pkg/token"
)

func GetClaims(ctx context.Context) (*token.Claims, error) {
	claims, ok := ctx.Value(interceptor.ContextKey{}).(*token.Claims)
	if !ok {
		return nil, token.ErrInvalidClaims
	}
	return claims, nil
}
