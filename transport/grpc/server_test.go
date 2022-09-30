package grpc

import (
	"context"
	"github.com/sado0823/go-kitx/internal/test/pbhelloworld"
	"github.com/sado0823/go-kitx/kit/middleware"
	"github.com/sado0823/go-kitx/transport"
	"google.golang.org/grpc"
	"testing"
)

type testHelloServer struct {
	pbhelloworld.UnimplementedGreeterServer
}

func Test_Server_Start(t *testing.T) {
	//server := NewServer(WithServerAddress("0.0.0.0:9009"))
	//pbhelloworld.RegisterGreeterServer(server, &testHelloServer{})
	//
	//endpoint, err := server.Endpoint()
	//if err != nil {
	//	t.Fatal(err)
	//}
	//t.Log(endpoint)
	//
	//go func() {
	//	err = server.Start(context.Background())
	//	if err != nil {
	//		panic(err)
	//	}
	//}()
	//
	//time.Sleep(time.Second * 1)
	//conn := testClient(t, server)
	//defer conn.Close()
	//
	//greeterClient := pbhelloworld.NewGreeterClient(conn)
	//for i := 0; i < 100; i++ {
	//	time.Sleep(time.Millisecond * 200)
	//	hello, err := greeterClient.SayHello(context.Background(), &pbhelloworld.HelloRequest{Name: "iamhello"})
	//	if err != nil {
	//		t.Error(err)
	//	}
	//	t.Log(hello)
	//}

}

func testClient(t *testing.T, srv *Server) *grpc.ClientConn {
	u, err := srv.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	// new a gRPC client
	conn, err := DialInsecure(context.Background(),
		WithClientEndpoint(u.Host),
		WithClientDialOptions(grpc.WithBlock()),
		WithClientUnaryInterceptor(
			func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				return invoker(ctx, method, req, reply, cc, opts...)
			}),
		WithClientMiddleware(func(handler middleware.Handler) middleware.Handler {
			return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
				if tr, ok := transport.FromClientContext(ctx); ok {
					header := tr.RequestHeader()
					header.Set("x-md-trace", "2233")
				}
				return handler(ctx, req)
			}
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	return conn
}
