package grpcserver

import (
	"net"

	"github.com/modulix-systems/goose-talk/pkg/logger"
	"google.golang.org/grpc"
)

type Server struct {
	grpc.ServiceRegistrar

	log      logger.Interface
	server   *grpc.Server
	ServeErr chan error
	Port     string
}

func New(log logger.Interface, port string) *Server {
	gRPCServer := grpc.NewServer()
	errChan := make(chan error, 1)
	return &Server{log: log, server: gRPCServer, ServeErr: errChan, Port: port}
}

	func (s *Server) RegisterService(desc *grpc.ServiceDesc, impl any) {
		s.server.RegisterService(desc,impl)
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
