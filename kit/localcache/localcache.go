package localcache

import (
	"context"
	"sync"
	"time"

	"github.com/sado0823/go-kitx/kit/localcache/internal"

	"golang.org/x/sync/singleflight"
)

const (
	defaultName         = "proc"
	timingWheelSlots    = 300
	timingWheelInterval = time.Second
)

type (
	Cache struct {
		name   string
		lock   sync.Mutex
		data   map[string]interface{}
		expire time.Duration

		lru         internal.Lru
		timingWheel *internal.TimingWheel
		sf          *singleflight.Group
		stat        *internal.Stat
	}

	Option func(cache *Cache)
)

func WithName(name string) Option {
	return func(cache *Cache) {
		cache.name = name
	}
}

func WithLimit(limit int) Option {
	return func(cache *Cache) {
		cache.lru = internal.NewLru(limit, cache.onEvict)
	}
}

func New(expire time.Duration, opts ...Option) (cache *Cache, err error) {
	cache = &Cache{
		data:   make(map[string]interface{}),
		expire: expire,
		lru:    internal.NewNoneLru(),
		sf:     &singleflight.Group{},
	}

	for _, opt := range opts {
		opt(cache)
	}

	if len(cache.name) == 0 {
		cache.name = defaultName
	}

	cache.stat = internal.NewStat(cache.name, cache.size)

	var tw *internal.TimingWheel
	tw, err = internal.NewTimingWheel(timingWheelInterval, timingWheelSlots, func(key, value interface{}) {
		v, ok := key.(string)
		if !ok {
			return
		}

		cache.Del(context.Background(), v)
	})
	if err != nil {
		return nil, err
	}

	cache.timingWheel = tw
	return cache, nil
}

func (c *Cache) Take(ctx context.Context, key string, fetch func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	if val, ok := c.doGet(ctx, key); ok {
		c.stat.Hit()
		return val, nil
	}

	var fresh bool
	val, err, _ := c.sf.Do(key, func() (interface{}, error) {
		// double check
		if val, ok := c.doGet(ctx, key); ok {
			c.stat.Hit()
			return val, nil
		}

		v, err := fetch(ctx)
		if err != nil {
			return nil, err
		}

		fresh = true
		c.Set(ctx, key, v)
		return v, nil
	})
	if err != nil {
		return nil, err
	}

	if fresh {
		c.stat.Miss()
		return val, nil
	}

	c.stat.Hit()
	return val, nil
}

func (c *Cache) Set(_ context.Context, key string, value interface{}) {
	c.lock.Lock()
	_, ok := c.data[key]
	c.data[key] = value
	c.lru.Add(key)
	c.lock.Unlock()

	if ok {
		c.timingWheel.MoveTimer(key, c.expire)
	} else {
		c.timingWheel.SetTimer(key, value, c.expire)
	}
}

func (c *Cache) Get(ctx context.Context, key string) (value interface{}, ok bool) {
	value, ok = c.doGet(ctx, key)
	if ok {
		c.stat.Hit()
	} else {
		c.stat.Miss()
	}

	return value, ok
}

func (c *Cache) Del(_ context.Context, key string) {
	c.lock.Lock()
	delete(c.data, key)
	c.lru.Remove(key)
	c.lock.Unlock()

	// using chan
	c.timingWheel.RemoveTimer(key)
}

func (c *Cache) doGet(_ context.Context, key string) (value interface{}, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	value, ok = c.data[key]
	if ok {
		c.lru.Add(key)
	}

	return value, ok
}

func (c *Cache) onEvict(key string) {
	// already locked
	delete(c.data, key)
	c.timingWheel.RemoveTimer(key)
}

func (c *Cache) size() int {
	c.lock.Lock()
	defer c.lock.Unlock()
	return len(c.data)
}
