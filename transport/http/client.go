package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sado0823/go-kitx/transport"
	"github.com/sado0823/go-kitx/transport/pbchain"
)

type (
	ClientOption func(op *Client)

	Client struct {
		ctx context.Context

		tlsConf      *tls.Config
		timeout      time.Duration
		endpoint     string
		host         string
		scheme       string
		userAgent    string
		encoder      EncodeRequestFunc
		decoder      DecodeResponseFunc
		errorDecoder DecodeErrorFunc
		transport    http.RoundTripper
		pbchain      []pbchain.Middleware
		insecure     bool

		httpClient *http.Client
	}
)

func WithClientTransport(trans http.RoundTripper) ClientOption {
	return func(o *Client) {
		o.transport = trans
	}
}

func WithClientTimeout(d time.Duration) ClientOption {
	return func(o *Client) {
		o.timeout = d
	}
}

func WithClientUserAgent(ua string) ClientOption {
	return func(o *Client) {
		o.userAgent = ua
	}
}

func WithClientPBChain(m ...pbchain.Middleware) ClientOption {
	return func(o *Client) {
		o.pbchain = m
	}
}

func WithClientEndpoint(endpoint string) ClientOption {
	return func(o *Client) {
		o.endpoint = endpoint
	}
}

func WithClientRequestEncoder(encoder EncodeRequestFunc) ClientOption {
	return func(o *Client) {
		o.encoder = encoder
	}
}

func WithClientResponseDecoder(decoder DecodeResponseFunc) ClientOption {
	return func(o *Client) {
		o.decoder = decoder
	}
}

func WithClientErrorDecoder(errorDecoder DecodeErrorFunc) ClientOption {
	return func(o *Client) {
		o.errorDecoder = errorDecoder
	}
}

func WithClientTLSConfig(c *tls.Config) ClientOption {
	return func(o *Client) {
		o.tlsConf = c
	}
}

func NewClient(ctx context.Context, opts ...ClientOption) (*Client, error) {
	client := &Client{
		ctx:          ctx,
		timeout:      time.Second * 2,
		encoder:      RequestEncoder,
		decoder:      ResponseDecoder,
		errorDecoder: ErrorDecoder,
		transport:    http.DefaultTransport,
	}
	for _, opt := range opts {
		opt(client)
	}
	if client.tlsConf != nil {
		if tr, ok := client.transport.(*http.Transport); ok {
			tr.TLSClientConfig = client.tlsConf
		}
	}
	client.insecure = client.tlsConf == nil

	urlV, err := parseClientEndpoint(client.endpoint, client.insecure)
	if err != nil {
		return nil, err
	}
	client.host = urlV.Host
	client.scheme = urlV.Scheme
	client.httpClient = &http.Client{
		Timeout:   client.timeout,
		Transport: client.transport,
	}

	return client, nil
}

func parseClientEndpoint(endpoint string, insecure bool) (*url.URL, error) {
	if !strings.Contains(endpoint, "://") {
		if insecure {
			endpoint = "http://" + endpoint
		} else {
			endpoint = "https://" + endpoint
		}
	}

	return url.Parse(endpoint)
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) Invoke(ctx context.Context, method, path string, args, reply interface{}, opts ...CallOption) error {
	call := defaultCallInfo(path)
	for _, opt := range opts {
		if err := opt.before(&call); err != nil {
			return err
		}
	}

	var (
		contentTypeStr string
		body           io.Reader
	)

	if args != nil {
		encoder, err := c.encoder(ctx, call.contentType, args)
		if err != nil {
			return err
		}
		contentTypeStr = call.contentType
		body = bytes.NewReader(encoder)
	}

	url := fmt.Sprintf("%s://%s%s", c.scheme, c.host, path)
	request, err := http.NewRequest(strings.ToUpper(method), url, body)
	if err != nil {
		return err
	}
	if contentTypeStr != "" {
		request.Header.Set("Content-Type", contentTypeStr)
	}
	if c.userAgent != "" {
		request.Header.Set("User-Agent", c.userAgent)
	}

	ctx = transport.NewClientContext(ctx, &Transport{
		endpoint:     c.endpoint,
		operation:    call.operation,
		reqHeader:    headerCarrier(request.Header),
		request:      request,
		pathTemplate: call.pathTemplate,
	})

	return c.invoke(ctx, request, args, reply, call, opts...)
}

func (c *Client) invoke(ctx context.Context, req *http.Request, args, reply interface{}, call callInfo, opts ...CallOption) error {
	h := func(ctx context.Context, args interface{}) (interface{}, error) {
		resp, err := c.do(req.WithContext(ctx))
		if resp != nil {
			cs := csAttempt{res: resp}
			for _, o := range opts {
				o.after(&call, &cs)
			}
		}
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if err = c.decoder(ctx, resp, reply); err != nil {
			return nil, err
		}
		return reply, nil
	}

	if len(c.pbchain) > 0 {
		h = pbchain.Chain(c.pbchain...)(h)
	}

	_, err := h(ctx, args)
	return err
}

// Do send an HTTP request and decodes the body of response into target.
// returns an error (of type *Error) if the response status code is not 2xx.
func (c *Client) Do(req *http.Request, opts ...CallOption) (*http.Response, error) {
	call := defaultCallInfo(req.URL.Path)
	for _, opt := range opts {
		if err := opt.before(&call); err != nil {
			return nil, err
		}
	}

	return c.do(req)
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	err = c.errorDecoder(req.Context(), response)

	return response, err
}
