package authgrpc

import (
	"context"
	"strconv"

	"github.com/Artem09076/dp/backend/auth_service/internal/lib/jwt"
	auth1 "github.com/Artem09076/dp/backend/auth_service/proto/gen/auth"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Register(ctx context.Context, email string, name string, inn string, business_type string, role string, password string, deviceID, ipAddress string) (string, string, error)
	Login1(ctx context.Context, email, password, deviceID, ipAddress string) (string, string, error)
	Login2(ctx context.Context, inn, password, deviceID, ipAddress string) (string, string, error)
	RefreshToken(ctx context.Context, refreshToken, deviceID, ipAddress string) (string, string, error)
	Logout(ctx context.Context, userID uuid.UUID, deviceID, accessToken string) error
	ValidateAccessToken(ctx context.Context, tokenString string) (*jwt.AccessClaims, error)
}

type AuthServer struct {
	auth1.UnimplementedAuthServiceServer
	auth Auth
}

func RegisterAuth(gRPC *grpc.Server, auth Auth) {
	authServer := &AuthServer{auth: auth}
	auth1.RegisterAuthServiceServer(gRPC, authServer)
}

func extractDeviceID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "unknown"
	}

	deviceIDs := md.Get("x-device-id")
	if len(deviceIDs) > 0 && deviceIDs[0] != "" {
		return deviceIDs[0]
	}

	return "unknown"
}

func extractIPAddress(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	ips := md.Get("x-forwarded-for")
	if len(ips) > 0 && ips[0] != "" {
		return ips[0]
	}

	return ""
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
	deviceID := extractDeviceID(ctx)
	ipAddress := extractIPAddress(ctx)

	accessToken, refreshToken, err := s.auth.Register(ctx, req.GetEmail(), req.GetName(), req.GetInn(), req.GetBusinessType(), req.GetUserRole(), req.GetPassword(), deviceID, ipAddress)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to register")
	}
	return &auth1.AuthResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *auth1.LoginRequest) (*auth1.AuthResponse, error) {
	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password are required")
	}
	deviceID := extractDeviceID(ctx)
	ipAddress := extractIPAddress(ctx)

	var accessToken, refreshToken string
	var err error
	if req.GetInn() == "" && req.GetEmail() != "" {
		accessToken, refreshToken, err = s.auth.Login1(ctx, req.GetEmail(), req.GetPassword(), deviceID, ipAddress)
	}
	if req.GetInn() != "" && req.GetEmail() == "" {
		accessToken, refreshToken, err = s.auth.Login2(ctx, req.GetInn(), req.GetPassword(), deviceID, ipAddress)
	}
	if req.GetInn() == "" && req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email or inn number are required")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &auth1.AuthResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *AuthServer) RefreshToken(ctx context.Context, req *auth1.RefreshTokenRequest) (*auth1.AuthResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	deviceID := extractDeviceID(ctx)
	ipAddress := extractIPAddress(ctx)

	accessToken, refreshToken, err := s.auth.RefreshToken(ctx, req.GetRefreshToken(), deviceID, ipAddress)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &auth1.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *auth1.LogoutRequest) (*auth1.LogoutResponse, error) {

	accessToken, ok := ctx.Value("access_token").(string)
	if !ok || accessToken == "" {
		return nil, status.Error(codes.Unauthenticated, "token not found in context")
	}
	userIDValue := ctx.Value("user_id")
	if userIDValue == nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	var userID uuid.UUID
	var err error

	switch v := userIDValue.(type) {
	case uuid.UUID:
		userID = v
	case string:
		userID, err = uuid.Parse(v)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid user ID")
		}
	default:
		return nil, status.Error(codes.Unauthenticated, "invalid user ID type")
	}

	deviceID := extractDeviceID(ctx)

	err = s.auth.Logout(ctx, userID, deviceID, accessToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &auth1.LogoutResponse{Success: true}, nil
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
