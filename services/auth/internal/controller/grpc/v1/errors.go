package rpc_v1

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInternalError = status.Error(codes.Internal, "Internal error")
)
