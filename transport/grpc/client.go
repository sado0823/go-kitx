package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/sado0823/go-kitx/kit/middleware"
	"github.com/sado0823/go-kitx/kit/registry"
	"github.com/sado0823/go-kitx/transport"
	"github.com/sado0823/go-kitx/transport/grpc/balancer/p2c"
	_ "github.com/sado0823/go-kitx/transport/grpc/resolver/direct"
	"github.com/sado0823/go-kitx/transport/grpc/resolver/discovery"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcinsecure "google.golang.org/grpc/credentials/insecure"
	grpcmd "google.golang.org/grpc/metadata"
)

type (
	ClientOption func(o *client)

	client struct {
		endpoint   string
		tlsConfig  *tls.Config
		timeout    time.Duration
		discovery  registry.Discovery
		middleware []middleware.Middleware

		unaryInts []grpc.UnaryClientInterceptor
		grpcOpts  []grpc.DialOption
	}
)

func WithClientEndpoint(endpoint string) ClientOption {
	return func(o *client) {
		o.endpoint = endpoint
	}
}

func WithClientTimeout(timeout time.Duration) ClientOption {
	return func(o *client) {
		o.timeout = timeout
	}
}

func WithClientMiddleware(m ...middleware.Middleware) ClientOption {
	return func(o *client) {
		o.middleware = m
	}
}

func WithClientDiscovery(d registry.Discovery) ClientOption {
	return func(o *client) {
		o.discovery = d
	}
}

func WithClientTLSConfig(c *tls.Config) ClientOption {
	return func(o *client) {
		o.tlsConfig = c
	}
}

func WithClientUnaryInterceptor(in ...grpc.UnaryClientInterceptor) ClientOption {
	return func(o *client) {
		o.unaryInts = in
	}
}

func WithClientDialOptions(opts ...grpc.DialOption) ClientOption {
	return func(o *client) {
		o.grpcOpts = opts
	}
}

func Dial(ctx context.Context, opts ...ClientOption) (*grpc.ClientConn, error) {
	return dial(ctx, false, opts...)
}

func DialInsecure(ctx context.Context, opts ...ClientOption) (*grpc.ClientConn, error) {
	return dial(ctx, true, opts...)
}

func dial(ctx context.Context, insecure bool, opts ...ClientOption) (*grpc.ClientConn, error) {
	opt := &client{
		timeout: time.Second * 2,
	}

	for _, op := range opts {
		op(opt)
	}

	ints := []grpc.UnaryClientInterceptor{
		unaryClientInterceptor(opt.middleware, opt.timeout),
	}
	ints = append(ints, opt.unaryInts...)

	grpcOpts := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingConfig": [{"%s":{}}]}`, p2c.Name)),
		grpc.WithChainUnaryInterceptor(ints...),
	}

	if opt.discovery != nil {
		resolver := grpc.WithResolvers(discovery.NewBuilder(opt.discovery, discovery.WithInsecure(insecure), discovery.WithTimeout(opt.timeout)))
		grpcOpts = append(grpcOpts, resolver)
	}

	if insecure {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(grpcinsecure.NewCredentials()))
	}

	if opt.tlsConfig != nil {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(opt.tlsConfig)))
	}

	grpcOpts = append(grpcOpts, opt.grpcOpts...)

	return grpc.DialContext(ctx, opt.endpoint, grpcOpts...)
}

func unaryClientInterceptor(ms []middleware.Middleware, timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = transport.NewClientContext(ctx, &Transport{
			endpoint:  cc.Target(),
			operation: method,
			reqHeader: headerCarrier{},
		})
		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			if tr, ok := transport.FromClientContext(ctx); ok {
				header := tr.RequestHeader()
				keys := header.Keys()
				keyvals := make([]string, 0, len(keys))
				for _, k := range keys {
					keyvals = append(keyvals, k, header.Get(k))
				}
				ctx = grpcmd.AppendToOutgoingContext(ctx, keyvals...)
			}
			return reply, invoker(ctx, method, req, reply, cc, opts...)
		}
		if len(ms) > 0 {
			h = middleware.Chain(ms...)(h)
		}
		_, err := h(ctx, req)
		return err
	}
}
