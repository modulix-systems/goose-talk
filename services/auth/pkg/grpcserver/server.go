package grpcserver

import (
	"fmt"
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
	s.server.RegisterService(desc, impl)
}

func (s *Server) Run() {
	listener, err := net.Listen("tcp", ":"+s.Port)
	if err != nil {
		s.log.Fatal(fmt.Errorf("grpcserver - Run - net.Listen: %w", err), "port", s.Port)
	}
	s.log.Info("gRPC server is ready to accept incoming requests", "address", listener.Addr().String())
	if err := s.server.Serve(listener); err != nil {
		s.log.Error(fmt.Errorf("Serve grpc server error: %w", err))
		s.ServeErr <- err
	}
}

func (s *Server) Stop() {
	s.log.Info("Stopping gRPC server")
	s.server.GracefulStop()
}
