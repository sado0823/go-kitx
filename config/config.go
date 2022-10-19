package config

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/sado0823/go-kitx/pkg/reflectx"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type (
	Configer interface {
		Reader
		Get(key string) interface{}
	}

	Reader interface {
		Load() error
		Scan(v interface{}) error
	}

	Option func(p *proxy)

	proxy struct {
		lock   sync.RWMutex
		kv     map[string]interface{}
		reader Reader
	}
)

func WithReader(reader Reader) Option {
	return func(p *proxy) {
		p.reader = reader
	}
}

func New(opts ...Option) Configer {
	p := &proxy{
		kv: make(map[string]interface{}),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *proxy) Load() error {
	p.lock.Lock()
	defer p.lock.Unlock()
	if err := p.reader.Load(); err != nil {
		return err
	}

	return p.reader.Scan(&p.kv)
}

func (p *proxy) Scan(v interface{}) error {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if m, ok := v.(proto.Message); ok {
		bytes, err := json.Marshal(p.kv)
		if err != nil {
			return err
		}
		return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(bytes, m)
	}
	return p.reader.Scan(v)
}

func (p *proxy) Get(key string) interface{} {
	v, _ := reflectx.PathSelect(context.Background(), key, p.kv)
	return v
}
