package authgrpc

import (
	"context"
	"strconv"

	auth1 "github.com/Artem09076/dp/backend/auth_service/proto/gen/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Register(ctx context.Context, email string, name string, inn string, business_type string, role string, password string) (string, error)
	Login1(ctx context.Context, email string, password string) (string, error)
	Login2(ctx context.Context, inn string, password string) (string, error)
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
	if req.UserRole == "performer" {
		if req.GetInn() == "" {
			return nil, status.Error(codes.InvalidArgument, "Inn is required")
		}
		if !s.isInnValid(req.GetInn()) {
			return nil, status.Error(codes.InvalidArgument, "Invalid Inn")
		}
		if req.GetBusinessType() == "" {
			return nil, status.Error(codes.InvalidArgument, "buisness type is required")
		}
	}
	accessToken, err := s.auth.Register(ctx, req.GetEmail(), req.GetName(), req.GetInn(), req.GetBusinessType(), req.GetUserRole(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to register")
	}
	return &auth1.AuthResponse{AccessToken: accessToken}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *auth1.LoginRequest) (*auth1.AuthResponse, error) {
	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password are required")
	}

	var accessToken string
	var err error
	if req.GetInn() == "" && req.GetEmail() != "" {
		accessToken, err = s.auth.Login1(ctx, req.GetEmail(), req.GetPassword())
	}
	if req.GetInn() != "" && req.GetEmail() == "" {
		accessToken, err = s.auth.Login2(ctx, req.GetInn(), req.GetPassword())
	}
	if req.GetInn() == "" && req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email or inn number are required")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &auth1.AuthResponse{AccessToken: accessToken}, nil
}

func (s *AuthServer) isInnValid(inn string) bool {
	if len(inn) != 12 {
		return false
	}
	w11 := []int{7, 2, 4, 10, 3, 5, 9, 4, 6, 8}
	w12 := []int{3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8}
	digit11 := s.getCheckDigit(w11, inn[:10])
	if string(inn[10]) != strconv.Itoa(digit11) {
		return false
	}

	digit12 := s.getCheckDigit(w12, inn[:11])
	if string(inn[11]) != strconv.Itoa(digit12) {
		return false
	}
	return true

}

func (s *AuthServer) getCheckDigit(weights []int, digitsStr string) int {
	sum := 0
	for i, w := range weights {
		digit, _ := strconv.Atoi(string(digitsStr[i]))
		sum += digit * w
	}
	return (sum % 11) % 10
}
