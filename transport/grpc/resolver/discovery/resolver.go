package discovery

import (
	"context"
	"errors"
	"time"

	"github.com/sado0823/go-kitx/kit/log"
	"github.com/sado0823/go-kitx/kit/registry"
	"github.com/sado0823/go-kitx/transport/internal/endpoint"

	"google.golang.org/grpc/attributes"
	googleResolver "google.golang.org/grpc/resolver"
)

type resolver struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	watcher    registry.Watcher
	conn       googleResolver.ClientConn

	insecure bool
}

func (r *resolver) ResolveNow(googleResolver.ResolveNowOptions) {

}

func (r *resolver) Close() {
	r.cancelFunc()
	err := r.watcher.Stop()
	if err != nil {
		log.Errorf("discovery resolver watcher stop err:%+v", err)
	}
}

func (r *resolver) watch() {
	for {
		select {
		case <-r.ctx.Done():
			return
		default:
		}
		svcs, err := r.watcher.Next()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Errorf("discovery resolver watch err:%+v", err)
			time.Sleep(time.Second)
			continue
		}
		r.update(svcs)
	}
}

func (r *resolver) update(svcs []*registry.Service) {
	addrs := make([]googleResolver.Address, 0)
	endpoints := make(map[string]struct{})
	for _, svc := range svcs {
		parseEndpoint, err := endpoint.ParseEndpoint(svc.Endpoints, endpoint.Scheme("grpc", !r.insecure))
		if err != nil {
			log.Errorf("discovery resolver update err:%+v", err)
			continue
		}

		if parseEndpoint == "" {
			continue
		}
		if _, ok := endpoints[parseEndpoint]; ok {
			continue
		}
		endpoints[parseEndpoint] = struct{}{}
		addr := googleResolver.Address{
			Addr:       parseEndpoint,
			ServerName: svc.Name,
			Attributes: parseAttributes(svc.Metadata),
		}
		addr.Attributes = addr.Attributes.WithValue("rawServiceInstance", r.insecure)
		addrs = append(addrs, addr)
	}
	if len(addrs) == 0 {
		log.Warnf("discovery resolver update found no addr, svcs:%#v", svcs)
		return
	}

	err := r.conn.UpdateState(googleResolver.State{
		Addresses: addrs,
	})

	if err != nil {
		log.Errorf("discovery resolver update err:%+v, svcs:%#v", err, svcs)
	}
}

func parseAttributes(md map[string]string) *attributes.Attributes {
	var a *attributes.Attributes
	for k, v := range md {
		if a == nil {
			a = attributes.New(k, v)
		} else {
			a = a.WithValue(k, v)
		}
	}
	return a
}
