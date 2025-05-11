package grpcserver

import (
	"net"

	"github.com/AlexeyTarasov77/talka-chats/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
)

type Server struct {
	log           logger.Interface
	server        *grpc.Server
	ServeErr      chan error
	healthChecker *health.Server
	Port          string
}

const fullSystemHealthServing = ""

func New(log *logger.Interface, port string, authServer ssov1.AuthServer, permissionsServer ssov1.PermissionsServer) *Server {
	gRPCServer := grpc.NewServer()
	ssov1.RegisterAuthServer(gRPCServer, authServer)
	ssov1.RegisterPermissionsServer(gRPCServer, permissionsServer)
	healthcheckServer := health.NewServer()
	healthgrpc.RegisterHealthServer(gRPCServer, healthcheckServer)
	return &Server{log, gRPCServer, make(chan error, 1), healthcheckServer, host, port}
}

func (self *Server) Run() {
	listener, err := net.Listen("tcp", ":"+self.Port)
	if err != nil {
		self.log.Error("Failed to listen", "error", err, "port", self.Port)
		panic(err)
	}
	self.log.Info("Starting gRPC server", "address", listener.Addr())
	self.healthChecker.SetServingStatus(fullSystemHealthServing, healthgrpc.HealthCheckResponse_SERVING)
	if err := self.server.Serve(listener); err != nil {
		self.log.Error("Failed to serve gRPC server", "error", err)
		self.healthChecker.SetServingStatus(fullSystemHealthServing, healthgrpc.HealthCheckResponse_NOT_SERVING)
		self.ServeErr <- err
	}
}

func (app *Server) Stop() {
	app.log.Info("Stopping gRPC server")
	app.healthChecker.SetServingStatus(fullSystemHealthServing, healthgrpc.HealthCheckResponse_NOT_SERVING)
	app.server.GracefulStop()
}
