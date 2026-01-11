package rpc_v1

import (
	pb "buf.build/gen/go/co3n/goose-proto/grpc/go/auth/v1/authv1grpc"
	"github.com/go-playground/validator/v10"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/modulix-systems/goose-talk/logger"
	"google.golang.org/grpc"
)

func Register(
	registrar grpc.ServiceRegistrar,
	authService *auth.AuthService,
	log logger.Interface,
	validate *validator.Validate,
) {
	auth := newAuthController(authService, log, validate)
	pb.RegisterAuthServiceServer(registrar, auth)
}
