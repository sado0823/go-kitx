package kitx

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sado0823/go-kitx/kit/registry"
	"github.com/sado0823/go-kitx/transport"

	"golang.org/x/sync/errgroup"
)

type (
	AppI interface {
		ID() string
		Name() string
		Version() string
		Metadata() map[string]string
		Endpoint() []string
	}

	App struct {
		ctxWithCancel context.Context
		opt           *option
		ctxCancel     context.CancelFunc
		lock          sync.Mutex
		registrySvc   *registry.Service
	}

	Option func(o *option)

	appKey struct{}
)

func NewContext(ctx context.Context, s AppI) context.Context {
	return context.WithValue(ctx, appKey{}, s)
}

func FromContext(ctx context.Context) (s AppI, ok bool) {
	s, ok = ctx.Value(appKey{}).(AppI)
	return
}

func New(opts ...Option) *App {
	opt := &option{
		ctx:              context.Background(),
		signals:          []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT},
		registrarTimeout: time.Second * 5,
		stopTimeout:      time.Second * 10,
		servers:          nil,
	}
	for _, o := range opts {
		o(opt)
	}

	opt.Fix()

	ctx, cancelFunc := context.WithCancel(opt.ctx)

	return &App{
		ctxWithCancel: ctx,
		opt:           opt,
		ctxCancel:     cancelFunc,
	}
}

func (a *App) ID() string {
	return a.opt.id
}

func (a *App) Name() string {
	return a.opt.name
}

func (a *App) Version() string {
	return a.opt.version
}

func (a *App) Metadata() map[string]string {
	return a.opt.metadata
}

func (a *App) Endpoint() []string {
	if a.registrySvc != nil {
		return a.registrySvc.Endpoints
	}
	return []string{}
}

func (a *App) Stop() error {
	a.lock.Lock()
	svc := a.registrySvc
	a.lock.Unlock()

	if a.opt.registrar != nil && svc != nil {
		ctx, cancel := context.WithTimeout(NewContext(a.ctxWithCancel, a), a.opt.registrarTimeout)
		defer cancel()
		if err := a.opt.registrar.Deregister(ctx, svc); err != nil {
			return err
		}
	}

	if a.ctxCancel != nil {
		a.ctxCancel()
	}

	return nil
}

func (a *App) Run() error {
	registrySvc, err := a.genRegistrySvc()
	if err != nil {
		return err
	}

	a.lock.Lock()
	a.registrySvc = registrySvc
	a.lock.Unlock()

	startingPrint(a.ID(), a.Name())

	var (
		eg, ctxWithApp = errgroup.WithContext(NewContext(a.ctxWithCancel, a))
		wg             sync.WaitGroup
	)

	for _, server := range a.opt.servers {
		server := server
		// server stop go-routine
		eg.Go(func() error {
			<-ctxWithApp.Done() // ctx will be canceled when app stop
			stopCtx, stopCancel := context.WithTimeout(NewContext(a.opt.ctx, a), a.opt.stopTimeout)
			defer stopCancel()
			return server.Stop(stopCtx)
		})
		wg.Add(1)
		// server start go-routine
		eg.Go(func() error {
			wg.Done()
			return server.Start(NewContext(a.opt.ctx, a))
		})
	}
	// wait all server started
	wg.Wait()

	// use registry
	if a.opt.registrar != nil {
		regisCtx, regisCancel := context.WithTimeout(ctxWithApp, a.opt.registrarTimeout)
		defer regisCancel()
		return a.opt.registrar.Register(regisCtx, a.registrySvc)
	}

	// wait signals for stop
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, a.opt.signals...)
	eg.Go(func() error {
		select {
		case <-ctxWithApp.Done():
			return nil
		case <-sig:
			return a.Stop()
		}
	})

	err = eg.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

func (a *App) genRegistrySvc() (*registry.Service, error) {
	endpoints := make([]string, 0, len(a.opt.endpoints))
	for _, endpoint := range a.opt.endpoints {
		endpoints = append(endpoints, endpoint.String())
	}
	if len(endpoints) == 0 {
		for _, server := range a.opt.servers {
			if endpoint, ok := server.(transport.Endpointer); ok {
				url, err := endpoint.Endpoint()
				if err != nil {
					return nil, err
				}
				endpoints = append(endpoints, url.String())
			}
		}
	}
	return &registry.Service{
		ID:        a.opt.id,
		Name:      a.opt.name,
		Version:   a.opt.version,
		Metadata:  a.opt.metadata,
		Endpoints: endpoints,
	}, nil
}
