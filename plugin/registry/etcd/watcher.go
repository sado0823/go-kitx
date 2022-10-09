package etcd

import (
	"context"

	"github.com/sado0823/go-kitx/kit/registry"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var _ registry.Watcher = (*watcher)(nil)

type watcher struct {
	key         string
	ctx         context.Context
	cancel      context.CancelFunc
	watchChan   clientv3.WatchChan
	watcher     clientv3.Watcher
	kv          clientv3.KV
	first       bool
	serviceName string
}

func newWatcher(ctx context.Context, prefix, name string, client *clientv3.Client) (*watcher, error) {
	w := &watcher{
		key:         prefix,
		ctx:         ctx,
		watcher:     clientv3.NewWatcher(client),
		kv:          clientv3.NewKV(client),
		first:       true,
		serviceName: name,
	}

	w.ctx, w.cancel = context.WithCancel(ctx)
	w.watchChan = w.watcher.Watch(w.ctx, prefix,
		clientv3.WithPrefix(), clientv3.WithRev(0), clientv3.WithKeysOnly())
	err := w.watcher.RequestProgress(context.Background())
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *watcher) Next() ([]*registry.Service, error) {
	if w.first {
		services, err := w.getInstances()
		w.first = false
		return services, err
	}

	select {
	case <-w.ctx.Done():
		return nil, w.ctx.Err()
	case <-w.watchChan:
		return w.getInstances()
	}
}

func (w *watcher) getInstances() ([]*registry.Service, error) {
	resp, err := w.kv.Get(w.ctx, w.key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	svcs := make([]*registry.Service, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		si, err := unmarshal(kv.Value)
		if err != nil {
			return nil, err
		}
		if si.Name != w.serviceName {
			continue
		}
		svcs = append(svcs, si)
	}

	return svcs, nil
}

func (w *watcher) Stop() error {
	return nil
}
