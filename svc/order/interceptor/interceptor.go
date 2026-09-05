package interceptor

import (
	"context"
	"order/types"
	"pkg/token"
	"uuid"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to parse userd id")
		}
		ctx = context.WithValue(ctx, types.UserIDKey{}, userID)
		return handler(ctx, req)
	}
}
