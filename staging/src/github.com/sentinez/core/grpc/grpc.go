package coregrpc

import (
	"context"

	grpcserver "github.com/sentinez/core/grpc/server"
	confpb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/conf/v1"
	typepb "github.com/sentinez/sentinez/api/gen/go/sentinez/types/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

// ServiceServer is a gRPC service server.
type ServiceServer interface {
	AsServer() *grpc.Server
	Serve(conf *confpb.Config) error
	Shutdown(ctx context.Context) error
}

// Server is a gRPC server that registers services.
type Server struct {
	server *grpc.Server
	meta   *typepb.XMeta
}

// Shutdown implements ServiceServer.
func (s *Server) Shutdown(_ context.Context) error {
	s.server.GracefulStop()
	return nil
}

// AsServer returns the underlying gRPC server.
// return the underlying gRPC server.
func (s *Server) AsServer() *grpc.Server {
	return s.server
}

// Serve starts the http server.
// return error if the http server fails to start.
func (s *Server) Serve(conf *confpb.Config) error {

	addr := conf.GetEnv().GetGrpcAddress()
	listener, err := grpcserver.ListenNetworkTCP(addr)
	if err != nil {
		return err
	}

	go Register(s.meta.GetServiceKey(), conf.GetEnv())
	return s.AsServer().Serve(listener)
}

func (s *Server) BufServe(bufLis *bufconn.Listener) error {
	return s.AsServer().Serve(bufLis)
}

// New returns a new service registrar.
// opts are the gRPC server options.
func New(conf *confpb.Config, opts ...grpc.ServerOption) *Server {
	return &Server{
		server: grpc.NewServer(opts...),
		meta:   conf.GetMeta(),
	}
}

// NewDefault returns a new service registrar with default options.
func NewDefault(conf *confpb.Config) *Server {
	return New(conf)
}

func NewDefaultServer() *Server {
	return &Server{
		server: grpc.NewServer(),
	}
}
