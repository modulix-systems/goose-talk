package grpcserver

import (
	"net"

	pb "buf.build/gen/go/co3n/goose-proto/grpc/go/auth/v1/authv1grpc"
	"github.com/modulix-systems/goose-talk/pkg/logger"
	"google.golang.org/grpc"
)

type Server struct {
	log      logger.Interface
	server   *grpc.Server
	ServeErr chan error
	Port     string
}

func New(log logger.Interface, port string, authServer pb.AuthServiceServer) *Server {

	gRPCServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(gRPCServer, authServer)
	return &Server{log, gRPCServer, make(chan error, 1), port}
}

func (s *Server) Run() {
	listener, err := net.Listen("tcp", ":"+s.Port)
	if err != nil {
		s.log.Error("Failed to listen", "error", err, "port", s.Port)
		panic(err)
	}
	s.log.Info("Starting gRPC server", "address", listener.Addr())
	if err := s.server.Serve(listener); err != nil {
		s.log.Error("Failed to serve gRPC server", "error", err)
		s.ServeErr <- err
	}
}

func (s *Server) Stop() {
	s.log.Info("Stopping gRPC server")
	s.server.GracefulStop()
}
