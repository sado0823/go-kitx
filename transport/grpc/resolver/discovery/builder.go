package discovery

import (
	"context"
	"fmt"
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

//	NewBuilder example discovery://<authority>/name
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

	var (
		done                                 = make(chan struct{}, 1)
		ctx, cancelFunc                      = context.WithCancel(context.Background())
		ctxWithTimeout, ctxWithTimeoutCancel = context.WithTimeout(context.Background(), b.timeout)
		// target.URL.Path like /demo.app
		appName = strings.TrimPrefix(target.URL.Path, "/")
	)

	defer ctxWithTimeoutCancel()

	go func() {
		watcher, err := b.discovery.Watch(ctx, appName)
		watchRes.err = err
		watchRes.w = watcher
		close(done)
	}()

	var err error
	select {
	case <-done:
		err = watchRes.err
	case <-ctxWithTimeout.Done():
		err = fmt.Errorf("discovery create watcher overtime, err: %+v", ctxWithTimeout.Err())
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

	// when init app, make sure have right endpoint
	if err = r.firstCheck(ctxWithTimeout); err != nil {
		cancelFunc()
		return nil, fmt.Errorf("first check err: %+v, app name: %s", err, appName)
	}

	go r.watch()

	return r, nil
}

func (b *builder) Scheme() string {
	return Scheme
}
