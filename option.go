package kitx

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/sado0823/go-kitx/kit/log"
	"github.com/sado0823/go-kitx/kit/registry"
	"github.com/sado0823/go-kitx/transport"

	"github.com/google/uuid"
)

type (
	option struct {
		id               string
		name             string
		version          string
		metadata         map[string]string
		endpoints        []*url.URL
		ctx              context.Context
		signals          []os.Signal
		logger           log.Logger
		registrar        registry.Registrar
		registrarTimeout time.Duration
		stopTimeout      time.Duration
		servers          []transport.Server
	}
)

func (opt *option) Fix() {
	if opt.id == "" {
		if newUUID, err := uuid.NewUUID(); err == nil {
			opt.id = newUUID.String()
		} else {
			opt.id = fmt.Sprintf("kitx.app.id.%d", time.Now().UnixNano())
		}
	}

	if opt.name == "" {
		opt.name = "kitx.app.name.unknown"
	}

	if opt.logger != nil {
		log.SetGlobal(opt.logger)
	}

}

// WithID id should be unique
func WithID(id string) Option {
	return func(o *option) { o.id = id }
}

func WithName(name string) Option {
	return func(o *option) { o.name = name }
}

func WithVersion(version string) Option {
	return func(o *option) { o.version = version }
}

func WithMetadata(md map[string]string) Option {
	return func(o *option) { o.metadata = md }
}

func WithEndpoint(endpoints ...*url.URL) Option {
	return func(o *option) { o.endpoints = endpoints }
}

func WithContext(ctx context.Context) Option {
	return func(o *option) { o.ctx = ctx }
}

func WithLogger(logger log.Logger) Option {
	return func(o *option) { o.logger = logger }
}

func WithServer(srv ...transport.Server) Option {
	return func(o *option) { o.servers = srv }
}

// WithSignal with exit signals.
func WithSignal(sigs ...os.Signal) Option {
	return func(o *option) { o.signals = sigs }
}

// WithRegistrar with service registry.
func WithRegistrar(r registry.Registrar) Option {
	return func(o *option) { o.registrar = r }
}

// WithRegistrarTimeout with registrar timeout.
func WithRegistrarTimeout(t time.Duration) Option {
	return func(o *option) { o.registrarTimeout = t }
}

// WithStopTimeout with app stop timeout.
func WithStopTimeout(t time.Duration) Option {
	return func(o *option) { o.stopTimeout = t }
}
