package authgrpc

import (
	"context"

	auth1 "github.com/Artem09076/dp/backend/auth_service/proto/gen/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Register(ctx context.Context, email string, name string, inn string, business_type string, role string, password string) (string, error)
	Login(ctx context.Context, email string, password string) (string, error)
}

type AuthServer struct {
	auth1.UnimplementedAuthServiceServer
	auth Auth
}

func RegisterAuth(gRPC *grpc.Server, auth Auth) {
	authServer := &AuthServer{auth: auth}
	auth1.RegisterAuthServiceServer(gRPC, authServer)
}

func (s *AuthServer) Register(ctx context.Context, req *auth1.RegisterRequest) (*auth1.AuthResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}
	accessToken, err := s.auth.Register(ctx, req.GetEmail(), req.GetName(), req.GetInn(), req.GetBusinessType(), req.GetUserRole(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to register")
	}
	return &auth1.AuthResponse{AccessToken: accessToken}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *auth1.LoginRequest) (*auth1.AuthResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}
	accessToken, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to login")
	}
	return &auth1.AuthResponse{AccessToken: accessToken}, nil
}
