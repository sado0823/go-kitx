package breaker

import (
	"sync"

	"github.com/sado0823/go-kitx/pkg/stringx"
)

var (
	lock     sync.RWMutex
	breakers = make(map[string]Breaker)
)

type (
	Breaker interface {
		Allow() error
		MarkSuccess()
		MarkFail()
		DoWithAcceptable(req func() error, acceptable func(err error) bool) error
	}

	OptionFn func(*wrapBreaker)

	wrapBreaker struct {
		name string
		Breaker
	}
)

func WithName(name string) OptionFn {
	return func(breaker *wrapBreaker) {
		breaker.name = name
	}
}

func New(options ...OptionFn) Breaker {
	b := &wrapBreaker{Breaker: newGoogleSre()}

	for i := range options {
		options[i](b)
	}
	if b.name == "" {
		b.name = stringx.Rand(8)
	}

	return b
}

// Get a breaker with name, if existed, return old one
func Get(name string) Breaker {
	lock.RLock()
	b, ok := breakers[name]
	lock.RUnlock()
	if ok {
		return b
	}

	lock.Lock()
	b, ok = breakers[name]
	if !ok {
		b = New(WithName(name))
		breakers[name] = b
	}
	lock.Unlock()

	return b
}

// Except make this breaker doesn't work
func Except(name string) {
	lock.Lock()
	breakers[name] = &noob{}
	lock.Unlock()
}
