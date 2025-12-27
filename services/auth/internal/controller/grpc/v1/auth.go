package rpc_v1

import (
	"context"

	"buf.build/gen/go/co3n/goose-proto/grpc/go/auth/v1/authv1grpc"
	pb "buf.build/gen/go/co3n/goose-proto/protocolbuffers/go/auth/v1"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/modulix-systems/goose-talk/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthV1 struct {
	authv1grpc.UnimplementedAuthServiceServer

	service *auth.AuthService
	log  logger.Interface
}

func (*AuthV1) SignUp(context.Context, *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignUp not implemented")
}

func (*AuthV1) SendOTP(context.Context, *pb.SendOTPRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendOTP not implemented")
}

func (*AuthV1) SignIn(context.Context, *pb.SignInRequest) (*pb.SignInResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignIn not implemented")
}

func newAuthController(service *auth.AuthService, log logger.Interface) *AuthV1 {
	return &AuthV1{
		service: service,
		log:  log,
	}
}
