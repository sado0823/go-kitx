package etcd

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/sado0823/go-kitx/kit/log"
	"github.com/sado0823/go-kitx/kit/registry"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	_ registry.Registrar = (*Registry)(nil)
	_ registry.Discovery = (*Registry)(nil)
)

type (
	Option func(o *options)

	options struct {
		ctx       context.Context
		namespace string
		ttl       time.Duration
		maxRetry  int
	}
)

func Context(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

func Namespace(namespace string) Option {
	return func(o *options) {
		o.namespace = namespace
	}
}

func TTL(ttl time.Duration) Option {
	return func(o *options) {
		o.ttl = ttl
	}
}

func MaxRetry(maxRetry int) Option {
	return func(o *options) {
		o.maxRetry = maxRetry
	}
}

type Registry struct {
	opt    *options
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

func New(client *clientv3.Client, opts ...Option) *Registry {
	op := &options{
		ctx:       context.Background(),
		namespace: "/kitx-services",
		ttl:       time.Second * 15,
		maxRetry:  5,
	}
	for _, opt := range opts {
		opt(op)
	}
	return &Registry{
		opt:    op,
		client: client,
		kv:     clientv3.NewKV(client),
		lease:  nil,
	}

}

func (r *Registry) Register(ctx context.Context, svc *registry.Service) error {
	key := fmt.Sprintf("%s/%s/%s", r.opt.namespace, svc.Name, svc.ID)
	svcStr, err := marshal(svc)
	if err != nil {
		return err
	}
	if r.lease != nil {
		r.lease.Close()
	}
	r.lease = clientv3.NewLease(r.client)
	leaseID, err := r.setKvLease(ctx, key, svcStr)
	if err != nil {
		return err
	}

	go r.heartbeat(r.opt.ctx, leaseID, key, svcStr)
	return nil
}

func (r *Registry) setKvLease(ctx context.Context, key string, value string) (clientv3.LeaseID, error) {
	grant, err := r.client.Grant(ctx, int64(r.opt.ttl.Seconds()))
	if err != nil {
		return 0, err
	}
	_, err = r.client.Put(ctx, key, value, clientv3.WithLease(grant.ID))
	if err != nil {
		return 0, err
	}
	return grant.ID, nil
}

func (r *Registry) heartbeat(ctx context.Context, leaseID clientv3.LeaseID, key string, value string) {
	curLeaseID := leaseID
	keepAlive, err := r.client.KeepAlive(ctx, leaseID)
	if err != nil {
		curLeaseID = 0
	}
	rand.Seed(time.Now().Unix())

	for {
		if curLeaseID == 0 {
			var do []int
			for now := 0; now < r.opt.maxRetry; now++ {
				if ctx.Err() != nil {
					return
				}

				idChan := make(chan clientv3.LeaseID, 1)
				errChan := make(chan error, 1)
				withCancel, cancelFunc := context.WithCancel(ctx)
				go func() {
					defer cancelFunc()
					setKvLeaseID, errLease := r.setKvLease(withCancel, key, value)
					if errLease != nil {
						errChan <- errLease
					} else {
						idChan <- setKvLeaseID
					}
				}()

				select {
				case <-time.After(3 * time.Second):
					cancelFunc()
					continue
				case err := <-errChan:
					log.Errorf("heartbeat setKvLease err:%+v", err)
					continue
				case curLeaseID = <-idChan:
				}

				keepAlive, err = r.client.KeepAlive(ctx, curLeaseID)
				if err == nil {
					break
				}
				do = append(do, 1<<now)
				time.Sleep(time.Duration(do[rand.Intn(len(do))]) * time.Second)
			}
			if _, ok := <-keepAlive; !ok {
				// retry failed
				return
			}
		}

		select {
		case _, ok := <-keepAlive:
			if !ok {
				if ctx.Err() != nil {
					// channel closed due to context cancel
					return
				}
				// need to retry registration
				curLeaseID = 0
				continue
			}
		case <-r.opt.ctx.Done():
			return
		}

	}

}

func (r *Registry) Deregister(ctx context.Context, svc *registry.Service) error {
	defer func() {
		if r.lease != nil {
			r.lease.Close()
		}
	}()
	key := fmt.Sprintf("%s/%s/%s", r.opt.namespace, svc.Name, svc.ID)
	_, err := r.client.Delete(ctx, key)
	return err
}

func (r *Registry) Get(ctx context.Context, name string) ([]*registry.Service, error) {
	prefix := fmt.Sprintf("%s/%s", r.opt.namespace, name)
	get, err := r.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	svcs := make([]*registry.Service, 0, len(get.Kvs))
	for _, kv := range get.Kvs {
		service, err := unmarshal(kv.Value)
		if err != nil {
			return nil, err
		}
		if service.Name != name {
			continue
		}
		svcs = append(svcs, service)
	}

	return svcs, nil
}

func (r *Registry) Watch(ctx context.Context, name string) (registry.Watcher, error) {
	key := fmt.Sprintf("%s/%s", r.opt.namespace, name)
	return newWatcher(ctx, key, name, r.client)
}
