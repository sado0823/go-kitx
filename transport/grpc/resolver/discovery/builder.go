package discovery

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/sado0823/go-kitx/kit/registry"

	googleResolver "google.golang.org/grpc/resolver"
)

const (
	Scheme = "discovery"
)

type (
	Option func(*builder)

	builder struct {
		discovery registry.Discovery
		timeout   time.Duration
		insecure  bool
	}
)

func WithTimeout(timeout time.Duration) Option {
	return func(b *builder) {
		b.timeout = timeout
	}
}

func WithInsecure(insecure bool) Option {
	return func(b *builder) {
		b.insecure = insecure
	}
}

func NewBuilder(dis registry.Discovery, opts ...Option) googleResolver.Builder {
	b := &builder{
		discovery: dis,
		timeout:   time.Second * 10,
		insecure:  false,
	}

	for _, opt := range opts {
		opt(b)
	}
	return b
}

func (b *builder) Build(target googleResolver.Target, cc googleResolver.ClientConn, opts googleResolver.BuildOptions) (googleResolver.Resolver, error) {
	watchRes := &struct {
		err error
		w   registry.Watcher
	}{}

	done := make(chan struct{}, 1)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		watcher, err := b.discovery.Watch(ctx, strings.TrimPrefix(target.URL.Path, "/"))
		watchRes.err = err
		watchRes.w = watcher
		close(done)
	}()

	var err error
	select {
	case <-done:
		err = watchRes.err
	case <-time.After(b.timeout):
		err = errors.New("discovery create watcher overtime")
	}
	if err != nil {
		cancelFunc()
		return nil, err
	}

	r := &resolver{
		ctx:        ctx,
		cancelFunc: cancelFunc,
		watcher:    watchRes.w,
		conn:       cc,
		insecure:   b.insecure,
	}
	r.watch()

	return r, nil
}

func (b *builder) Scheme() string {
	return Scheme
}
