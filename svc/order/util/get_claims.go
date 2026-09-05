package util

import (
	"context"
	"order/types"
	"pkg/token"
)

func GetClaims(ctx context.Context) (*token.Claims, error) {
	claims, ok := ctx.Value(types.UserIDKey{}).(*token.Claims)
	if !ok {
		return nil, token.ErrInvalidClaims
	}
	return claims, nil
}
