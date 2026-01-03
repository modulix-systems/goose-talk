package rpc_v1

import (
	"context"
	"errors"
	"fmt"

	"buf.build/gen/go/co3n/goose-proto/grpc/go/auth/v1/authv1grpc"
	pb "buf.build/gen/go/co3n/goose-proto/protocolbuffers/go/auth/v1"
	"github.com/go-playground/validator/v10"
	"github.com/modulix-systems/goose-talk/internal/dtos"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/modulix-systems/goose-talk/pkg/logger"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthV1 struct {
	authv1grpc.UnimplementedAuthServiceServer

	service  *auth.AuthService
	log      logger.Interface
	validate *validator.Validate
}

func (a *AuthV1) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	reqDto := &dtos.SignUpRequest{
		Username:         req.GetUsername(),
		Password:         req.Password,
		Email:            req.GetEmail(),
		FirstName:        req.GetFirstName(),
		LastName:         req.GetLastName(),
		ConfirmationCode: req.GetIpAddr(),
		IpAddr:           req.GetIpAddr(),
		DeviceInfo:       req.GetDeviceInfo(),
		BirthDate:        req.BirthDate.AsTime(),
		AboutMe:          req.GetAboutMe(),
	}
	if errs := reqDto.Validate(); errs != nil {
		st := status.New(codes.InvalidArgument, "Validation error")
		st, _ = st.WithDetails(&errdetails.BadRequest{FieldViolations: errs})
		return nil, st.Err()
	}
	result, err := a.service.SignUp(ctx, reqDto)
	if err != nil {
		if errors.Is(err, auth.ErrOtpIsNotValid) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &pb.SignUpResponse{
		User:    mapUser(result.User),
		Session: mapSession(result.Session),
	}, nil
}

func (a *AuthV1) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.SignInResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignIn not implemented")
}

func newAuthController(service *auth.AuthService, log logger.Interface, validate *validator.Validate) *AuthV1 {
	return &AuthV1{
		service:  service,
		log:      log,
		validate: validate,
	}
}
