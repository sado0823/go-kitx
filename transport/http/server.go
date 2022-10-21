package http

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/sado0823/go-kitx/internal/host"
	"github.com/sado0823/go-kitx/kit/log"
	"github.com/sado0823/go-kitx/transport"
	"github.com/sado0823/go-kitx/transport/internal/endpoint"
	"github.com/sado0823/go-kitx/transport/pbchain"

	"github.com/gorilla/mux"
)

var (
	_ transport.Server     = (*Server)(nil)
	_ transport.Endpointer = (*Server)(nil)
	_ http.Handler         = (*Server)(nil)
)

type (
	ServerOption func(op *Server)

	Server struct {
		*http.Server
		router   *mux.Router
		listener net.Listener

		strictSlash bool
		tlsConfig   *tls.Config
		endpoint    *url.URL
		network     string
		address     string
		timeout     time.Duration

		filters []FilterFunc
		pbchain pbchain.Matcher

		reqDecoder   DecodeRequestFunc
		respEncoder  EncodeResponseFunc
		errorEncoder EncodeErrorFunc

		err error
	}
)

func WithServerNetwork(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

func WithServerAddress(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

func WithServerTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

func WithServerPBChain(m ...pbchain.Middleware) ServerOption {
	return func(o *Server) {
		o.pbchain.Use(m...)
	}
}

func WithServerFilter(filters ...FilterFunc) ServerOption {
	return func(o *Server) {
		o.filters = filters
	}
}

func WithServerRequestDecoder(dec DecodeRequestFunc) ServerOption {
	return func(o *Server) {
		o.reqDecoder = dec
	}
}

func WithServerResponseEncoder(en EncodeResponseFunc) ServerOption {
	return func(o *Server) {
		o.respEncoder = en
	}
}

func WithServerErrorEncoder(en EncodeErrorFunc) ServerOption {
	return func(o *Server) {
		o.errorEncoder = en
	}
}

func WithServerTLSConfig(c *tls.Config) ServerOption {
	return func(o *Server) {
		o.tlsConfig = c
	}
}

// WithServerStrictSlash is with mux's StrictSlash
// If true, when the path pattern is "/path/", accessing "/path" will
// redirect to the former and vice versa.
func WithServerStrictSlash(strictSlash bool) ServerOption {
	return func(o *Server) {
		o.strictSlash = strictSlash
	}
}

func WithServerListener(lis net.Listener) ServerOption {
	return func(s *Server) {
		s.listener = lis
	}
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		strictSlash:  true,
		network:      "tcp",
		address:      ":0",
		timeout:      time.Second * 2,
		pbchain:      pbchain.NewMatcher(),
		reqDecoder:   RequestDecoder,
		respEncoder:  ResponseEncoder,
		errorEncoder: ErrorEncoder,
	}

	for _, opt := range opts {
		opt(srv)
	}

	srv.router = mux.NewRouter().StrictSlash(srv.strictSlash)
	srv.router.NotFoundHandler = http.DefaultServeMux
	srv.router.MethodNotAllowedHandler = http.DefaultServeMux
	srv.router.Use(srv.baseFilter())
	srv.Server = &http.Server{
		Handler:   FilterChain(srv.filters...)(srv.router),
		TLSConfig: srv.tlsConfig,
	}

	return srv
}

func (s *Server) baseFilter() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			var (
				ctx    context.Context
				cancel context.CancelFunc
			)
			if s.timeout > 0 {
				ctx, cancel = context.WithTimeout(req.Context(), s.timeout)
			} else {
				ctx, cancel = context.WithCancel(req.Context())
			}
			defer cancel()

			pathTemplate := req.URL.Path
			if route := mux.CurrentRoute(req); route != nil {
				// /path/123 -> /path/{id}
				pathTemplate, _ = route.GetPathTemplate()
			}

			tr := &Transport{
				operation:    pathTemplate,
				pathTemplate: pathTemplate,
				reqHeader:    headerCarrier(req.Header),
				replyHeader:  headerCarrier(w.Header()),
				request:      req,
			}
			if s.endpoint != nil {
				tr.endpoint = s.endpoint.String()
			}
			tr.request = req.WithContext(transport.NewServerContext(ctx, tr))
			next.ServeHTTP(w, tr.request)
		})
	}
}

func (s *Server) Start(ctx context.Context) error {
	if err := s.listenAndEndpoint(); err != nil {
		return err
	}
	s.BaseContext = func(net.Listener) context.Context {
		return ctx
	}
	log.Infof("http server listen on: %s", s.listener.Addr())
	var err error
	if s.tlsConfig != nil {
		err = s.ServeTLS(s.listener, "", "")
	} else {
		err = s.Serve(s.listener)
	}

	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	log.Info("http server stop")
	return s.Shutdown(ctx)
}

// Endpoint return a real address to registry endpoint.
// examples:
//
//	https://127.0.0.1:8000
//	Legacy: http://127.0.0.1:8000?isSecure=false
func (s *Server) Endpoint() (*url.URL, error) {
	if err := s.listenAndEndpoint(); err != nil {
		return nil, err
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
		s.endpoint = endpoint.NewEndpoint(endpoint.Scheme("http", s.tlsConfig != nil), addr)
	}
	return s.err
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.Handler.ServeHTTP(w, req)
}

// Use uses a service pbchain with selector.
// selector:
//   - '/*'
//   - '/foo/*'
//   - '/foo/bar'
func (s *Server) Use(selector string, m ...pbchain.Middleware) {
	s.pbchain.Add(selector, m...)
}

// WalkRoute walks the router and all its sub-routers, calling walkFn for each route in the tree.
func (s *Server) WalkRoute(fn WalkRouteFunc) error {
	return s.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		methods, err := route.GetMethods()
		if err != nil {
			return nil // ignore no methods
		}
		path, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		for _, method := range methods {
			if err := fn(RouteInfo{Method: method, Path: path}); err != nil {
				return err
			}
		}
		return nil
	})
}

// Route registers an HTTP router.
func (s *Server) Route(prefix string, filters ...FilterFunc) *Router {
	return newRouter(prefix, s, filters...)
}

// Handle registers a new route with a matcher for the URL path.
func (s *Server) Handle(path string, h http.Handler) {
	s.router.Handle(path, h)
}

// HandlePrefix registers a new route with a matcher for the URL path prefix.
func (s *Server) HandlePrefix(prefix string, h http.Handler) {
	s.router.PathPrefix(prefix).Handler(h)
}

// HandleFunc registers a new route with a matcher for the URL path.
func (s *Server) HandleFunc(path string, h http.HandlerFunc) {
	s.router.HandleFunc(path, h)
}

// HandleHeader registers a new route with a matcher for the header.
func (s *Server) HandleHeader(key, val string, h http.HandlerFunc) {
	s.router.Headers(key, val).Handler(h)
}
