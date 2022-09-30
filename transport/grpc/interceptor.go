package grpc

import (
	"context"

	contextx "github.com/sado0823/go-kitx/internal/context"
	"github.com/sado0823/go-kitx/kit/middleware"
	"github.com/sado0823/go-kitx/transport"

	"google.golang.org/grpc"
	grpcmd "google.golang.org/grpc/metadata"
)

func (s *Server) defaultUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		mergeCtx, cancelFunc := contextx.Merge(ctx, s.ctx)
		defer cancelFunc()

		md, _ := grpcmd.FromIncomingContext(mergeCtx)
		replyHeader := grpcmd.MD{}

		tr := &Transport{
			operation:   info.FullMethod,
			reqHeader:   headerCarrier(md),
			replyHeader: headerCarrier(replyHeader),
		}
		if s.endpoint != nil {
			tr.endpoint = s.endpoint.String()
		}

		mergeCtx = transport.NewServerContext(mergeCtx, tr)
		if s.timeout > 0 {
			mergeCtx, cancelFunc = context.WithTimeout(ctx, s.timeout)
			defer cancelFunc()
		}

		handlerFn := func(ctx context.Context, req interface{}) (interface{}, error) {
			return handler(ctx, req)
		}

		if next := s.middleware.Match(tr.Operation()); len(next) > 0 {
			handlerFn = middleware.Chain(next...)(handlerFn)
		}

		reply, err := handlerFn(mergeCtx, req)
		if len(replyHeader) > 0 {
			_ = grpc.SetHeader(ctx, replyHeader)
		}

		return reply, err
	}
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func NewWrappedStream(ctx context.Context, stream grpc.ServerStream) grpc.ServerStream {
	return &wrappedStream{
		ServerStream: stream,
		ctx:          ctx,
	}
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

// streamServerInterceptor is a gRPC stream server interceptor
func (s *Server) defaultStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx, cancel := contextx.Merge(ss.Context(), s.ctx)
		defer cancel()
		md, _ := grpcmd.FromIncomingContext(ctx)
		replyHeader := grpcmd.MD{}
		ctx = transport.NewServerContext(ctx, &Transport{
			endpoint:    s.endpoint.String(),
			operation:   info.FullMethod,
			reqHeader:   headerCarrier(md),
			replyHeader: headerCarrier(replyHeader),
		})

		ws := NewWrappedStream(ctx, ss)

		err := handler(srv, ws)
		if len(replyHeader) > 0 {
			_ = grpc.SetHeader(ctx, replyHeader)
		}
		return err
	}
}
