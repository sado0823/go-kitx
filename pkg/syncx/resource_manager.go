package syncx

import (
	"io"
	"sync"

	"github.com/sado0823/go-kitx/pkg/errorx"

	"golang.org/x/sync/singleflight"
)

type ResourceManager struct {
	resources map[string]io.Closer
	sf        *singleflight.Group
	lock      sync.RWMutex
}

func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		resources: make(map[string]io.Closer),
		sf:        &singleflight.Group{},
	}
}

func (r *ResourceManager) Close() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	var batchErr errorx.Batch
	for _, resource := range r.resources {
		if err := resource.Close(); err != nil {
			batchErr.Add(err)
		}
	}

	r.resources = nil
	return batchErr.Err()
}

func (r *ResourceManager) Get(key string, creator func() (io.Closer, error)) (io.Closer, error) {
	do, err, _ := r.sf.Do(key, func() (interface{}, error) {
		r.lock.RLock()
		rsc, ok := r.resources[key]
		r.lock.RUnlock()
		if ok {
			return rsc, nil
		}

		rsc, err := creator()
		if err != nil {
			return nil, err
		}

		r.lock.Lock()
		defer r.lock.Unlock()
		r.resources[key] = rsc

		return rsc, nil
	})
	if err != nil {
		return nil, err
	}

	return do.(io.Closer), nil
}

func (r *ResourceManager) Inject(key string, rsc io.Closer) {
	r.lock.Lock()
	r.resources[key] = rsc
	r.lock.Unlock()
}
