package interceptor

import (
	"context"
	"pkg/token"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ContextKey struct{}

func TokenInterceptor(validator token.ITokenValidator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		carrier, ok := req.(TokenCarrier)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "request does not implement TokenCarrier")
		}
		tokenStr := carrier.GetToken()
		claims, err := validator.GetClaims(tokenStr, token.Access)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		ctx = context.WithValue(ctx, ContextKey{}, claims)
		return handler(ctx, req)
	}
}
