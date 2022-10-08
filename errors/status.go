package errors

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

const (
	// CodeClientClosed is non-standard http status code,
	// which defined by nginx.
	// https://httpstatus.in/499/
	CodeClientClosed = 499
)

type (
	Converter interface {
		HTTP2GRPC(http int) codes.Code
		GRPC2HTTP(grpc codes.Code) int
	}

	statusCodeConverter struct{}
)

var (
	DefaultConverter Converter = statusCodeConverter{}

	http2grpc = map[int]codes.Code{
		http.StatusOK:                  codes.OK,
		http.StatusBadRequest:          codes.InvalidArgument,
		http.StatusUnauthorized:        codes.Unauthenticated,
		http.StatusForbidden:           codes.PermissionDenied,
		http.StatusNotFound:            codes.NotFound,
		http.StatusConflict:            codes.Aborted,
		http.StatusTooManyRequests:     codes.ResourceExhausted,
		http.StatusInternalServerError: codes.Internal,
		http.StatusNotImplemented:      codes.Unimplemented,
		http.StatusServiceUnavailable:  codes.Unavailable,
		http.StatusGatewayTimeout:      codes.DeadlineExceeded,
		CodeClientClosed:               codes.Canceled,
	}

	grpc2http = map[codes.Code]int{
		codes.OK:                http.StatusOK,
		codes.InvalidArgument:   http.StatusBadRequest,
		codes.Unauthenticated:   http.StatusUnauthorized,
		codes.PermissionDenied:  http.StatusForbidden,
		codes.NotFound:          http.StatusNotFound,
		codes.Aborted:           http.StatusConflict,
		codes.ResourceExhausted: http.StatusTooManyRequests,
		codes.Internal:          http.StatusInternalServerError,
		codes.Unimplemented:     http.StatusNotImplemented,
		codes.Unavailable:       http.StatusServiceUnavailable,
		codes.DeadlineExceeded:  http.StatusGatewayTimeout,
		codes.Canceled:          CodeClientClosed,
	}
)

func (s statusCodeConverter) HTTP2GRPC(http int) codes.Code {
	code, ok := http2grpc[http]
	if !ok {
		return codes.Unknown
	}
	return code
}

func (s statusCodeConverter) GRPC2HTTP(grpc codes.Code) int {
	code, ok := grpc2http[grpc]
	if !ok {
		return http.StatusInternalServerError
	}
	return code
}

// ToGRPCCode converts an HTTP error code into the corresponding gRPC response status.
func ToGRPCCode(code int) codes.Code {
	return DefaultConverter.HTTP2GRPC(code)
}

// ToHTTPCode converts a gRPC error code into the corresponding HTTP response status.
func ToHTTPCode(code codes.Code) int {
	return DefaultConverter.GRPC2HTTP(code)
}
