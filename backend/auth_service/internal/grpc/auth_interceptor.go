package authgrpc

import (
	"context"
	"strings"

	"github.com/Artem09076/dp/backend/auth_service/internal/lib/jwt"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type TokenValidator interface {
	ValidateAccessToken(ctx context.Context, tokenString string) (*jwt.AccessClaims, error)
}

func AuthInterceptor(validator TokenValidator, jwtSecret []byte) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		skipMethods := map[string]bool{
			"/auth.AuthService/Register":     true,
			"/auth.AuthService/Login":        true,
			"/auth.AuthService/RefreshToken": true,
		}

		if skipMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
		}

		authHeader := md["authorization"]
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
		}

		tokenString := strings.TrimPrefix(authHeader[0], "Bearer ")
		if tokenString == "" {
			return nil, status.Error(codes.Unauthenticated, "invalid token format")
		}

		claims, err := jwt.ParseAccessToken(tokenString, jwtSecret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid user ID in token")
		}

		ctx = context.WithValue(ctx, "user_id", userID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		ctx = context.WithValue(ctx, "user_role", claims.Role)
		ctx = context.WithValue(ctx, "device_id", claims.DeviceID)
		ctx = context.WithValue(ctx, "access_token", tokenString)

		return handler(ctx, req)
	}
}
