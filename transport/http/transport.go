package http

import (
	"context"
	"net/http"

	"github.com/sado0823/go-kitx/transport"
)

var _ Transporter = (*Transport)(nil)

type (
	Transporter interface {
		transport.Transporter
		Request() *http.Request
		PathTemplate() string
	}

	Transport struct {
		endpoint     string
		operation    string
		reqHeader    headerCarrier
		replyHeader  headerCarrier
		request      *http.Request
		pathTemplate string
	}
)

func SetOperation(ctx context.Context, op string) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		if tr, ok := tr.(*Transport); ok {
			tr.operation = op
		}
	}
}

func (t *Transport) Kind() transport.Kind {
	return transport.KindHTTP
}

func (t *Transport) Endpoint() string {
	return t.endpoint
}

func (t *Transport) Operation() string {
	return t.operation
}

func (t *Transport) RequestHeader() transport.Header {
	return t.reqHeader
}

func (t *Transport) ReplyHeader() transport.Header {
	return t.replyHeader
}

func (t *Transport) Request() *http.Request {
	return t.request
}

func (t *Transport) PathTemplate() string {
	return t.pathTemplate
}

type headerCarrier http.Header

// Get returns the value associated with the passed key.
func (hc headerCarrier) Get(key string) string {
	return http.Header(hc).Get(key)
}

// Set stores the key-value pair.
func (hc headerCarrier) Set(key string, value string) {
	http.Header(hc).Set(key, value)
}

// Keys lists the keys stored in this carrier.
func (hc headerCarrier) Keys() []string {
	keys := make([]string, 0, len(hc))
	for k := range http.Header(hc) {
		keys = append(keys, k)
	}
	return keys
}
