package grpc

import (
	"context"
	"crypto/tls"
	"net"
	"net/url"
	"time"

	"github.com/sado0823/go-kitx/internal/host"
	"github.com/sado0823/go-kitx/kit/log"
	"github.com/sado0823/go-kitx/transport"
	"github.com/sado0823/go-kitx/transport/internal/endpoint"
	"github.com/sado0823/go-kitx/transport/pbchain"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var (
	_ transport.Server     = (*Server)(nil)
	_ transport.Endpointer = (*Server)(nil)
)

type (
	ServerOption func(o *Server)

	Server struct {
		*grpc.Server
		ctx       context.Context
		tlsConfig *tls.Config
		listener  net.Listener
		err       error
		network   string
		address   string
		endpoint  *url.URL
		timeout   time.Duration

		pbchain    pbchain.Matcher
		unaryInts  []grpc.UnaryServerInterceptor
		streamInts []grpc.StreamServerInterceptor
		grpcOpts   []grpc.ServerOption
		health     *health.Server
	}
)

// WithServerPBChain with server pbchain.
func WithServerPBChain(m ...pbchain.Middleware) ServerOption {
	return func(s *Server) {
		s.pbchain.Use(m...)
	}
}

// WithServerNetwork with server network. default is `tcp`
func WithServerNetwork(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

// WithServerAddress with server address like `0.0.0.0:9000`. default is `:0`
func WithServerAddress(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

// WithServerTimeout with server timeout.
func WithServerTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

// WithServerTLSConfig with TLS config.
func WithServerTLSConfig(c *tls.Config) ServerOption {
	return func(s *Server) {
		s.tlsConfig = c
	}
}

// WithServerListener with server lis
func WithServerListener(lis net.Listener) ServerOption {
	return func(s *Server) {
		s.listener = lis
	}
}

// WithServerOption with grpc options.
func WithServerOption(opts ...grpc.ServerOption) ServerOption {
	return func(s *Server) {
		s.grpcOpts = opts
	}
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		ctx:     context.Background(),
		network: "tcp",
		address: ":0",
		timeout: time.Second * 3,
		health:  health.NewServer(),
		pbchain: pbchain.NewMatcher(),
	}
	for _, opt := range opts {
		opt(srv)
	}

	interUnaryInts := []grpc.UnaryServerInterceptor{
		srv.defaultUnaryServerInterceptor(),
	}

	interStreamInts := []grpc.StreamServerInterceptor{
		srv.defaultStreamServerInterceptor(),
	}

	interUnaryInts = append(interUnaryInts, srv.unaryInts...)
	interStreamInts = append(interStreamInts, srv.streamInts...)

	grpcOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(interUnaryInts...),
		grpc.ChainStreamInterceptor(interStreamInts...),
	}
	if srv.tlsConfig != nil {
		grpcOpts = append(grpcOpts, grpc.Creds(credentials.NewTLS(srv.tlsConfig)))
	}
	// merge grpc options
	if len(srv.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, srv.grpcOpts...)
	}

	srv.Server = grpc.NewServer(grpcOpts...)
	grpc_health_v1.RegisterHealthServer(srv.Server, srv.health)
	reflection.Register(srv.Server)
	return srv
}

func (s *Server) Start(ctx context.Context) error {
	err := s.listenAndEndpoint()
	if err != nil {
		return err
	}
	s.ctx = ctx
	log.Infof("grpc server listen on: %s", s.listener.Addr())
	s.health.Resume()
	return s.Serve(s.listener)
}

func (s *Server) Stop(ctx context.Context) error {
	s.health.Shutdown()
	s.GracefulStop()
	log.Info("grpc server stop")
	return nil
}

// Use uses a service pbchain with selector.
// selector:
//   - '/*'
//   - '/helloworld.v1.Greeter/*'
//   - '/helloworld.v1.Greeter/SayHello'
func (s *Server) Use(selector string, m ...pbchain.Middleware) {
	s.pbchain.Add(selector, m...)
}

// Endpoint return a real address to registry endpoint.
// examples:
//
//	grpc://127.0.0.1:9000?isSecure=false
func (s *Server) Endpoint() (*url.URL, error) {
	if err := s.listenAndEndpoint(); err != nil {
		return nil, s.err
	}
	return s.endpoint, nil
}

func (s *Server) listenAndEndpoint() error {
	if s.listener == nil {
		lis, err := net.Listen(s.network, s.address)
		if err != nil {
			s.err = err
			return err
		}
		s.listener = lis
	}
	if s.endpoint == nil {
		addr, err := host.Extract(s.address, s.listener)
		if err != nil {
			s.err = err
			return err
		}
		s.endpoint = endpoint.NewEndpoint(endpoint.Scheme("grpc", s.tlsConfig != nil), addr)
	}
	return s.err
}
