package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	authgrpc "github.com/Artem09076/dp/backend/auth_service/internal/grpc"
	authpb "github.com/Artem09076/dp/backend/auth_service/proto/gen/auth"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type App struct {
	log         *slog.Logger
	httpServer  *http.Server
	grpcPort    uint
	gatewayPort uint
	jwtSecret   []byte
	authService authgrpc.Auth
}

func New(log *slog.Logger, authService authgrpc.Auth, grpcPort, gatewayPort uint, jwtSecret []byte) *App {
	return &App{
		log:         log,
		authService: authService,
		grpcPort:    grpcPort,
		gatewayPort: gatewayPort,
		jwtSecret:   jwtSecret,
	}
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token != "" {
				ctx := context.WithValue(r.Context(), "access_token", token)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) Start() error {
	ctx := context.Background()

	mux := runtime.NewServeMux(
		runtime.WithMetadata(func(ctx context.Context, r *http.Request) metadata.MD {
			md := make(map[string]string)

			if deviceID := r.Header.Get("X-Device-ID"); deviceID != "" {
				md["x-device-id"] = deviceID
			}

			if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
				md["x-forwarded-for"] = ip
			}

			if authHeader := r.Header.Get("Authorization"); authHeader != "" {
				md["authorization"] = authHeader
			}

			if token, ok := ctx.Value("access_token").(string); ok && token != "" {
				md["access_token"] = token
			}

			return metadata.New(md)
		}),
	)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	err := authpb.RegisterAuthServiceHandlerFromEndpoint(
		ctx,
		mux,
		fmt.Sprintf("localhost:%d", a.grpcPort),
		opts,
	)
	if err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
	}

	handler := authMiddleware(mux)

	a.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", a.gatewayPort),
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	a.log.Info("starting HTTP gateway", "port", a.gatewayPort)
	return a.httpServer.ListenAndServe()
}

func (a *App) Stop(ctx context.Context) {
	a.log.Info("stopping HTTP gateway")
	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			a.log.Error("gateway shutdown error", "error", err)
		}
	}
}
