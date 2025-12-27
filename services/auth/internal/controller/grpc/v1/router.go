package rpc_v1

import (
	pb "buf.build/gen/go/co3n/goose-proto/grpc/go/auth/v1/authv1grpc"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/modulix-systems/goose-talk/pkg/logger"
	"google.golang.org/grpc"
)

func Register(
	s grpc.ServiceRegistrar,
	authService *auth.AuthService,
	log logger.Interface,
) {
	auth := newAuthController(authService, log)
	pb.RegisterAuthServiceServer(s, auth)
}
