package breaker

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"

	rollingwindow "github.com/sado0823/go-kitx/pkg/rollingwindow/v2"
)

var ErrGoogleSreBreakOn = errors.New("google sre breaker is on")

type (
	GoogleSreOptionFn func(*googleSreOption)

	googleSreOption struct {
		windowTime  time.Duration
		bucketCount int
		sreK        float64
		protection  int
	}

	GoogleSre struct {
		option *googleSreOption
		rw     *rollingwindow.RollingWindow
		lock   sync.Mutex
		rand   *rand.Rand
	}
)

func WithGoogleSreWindow(windowTime time.Duration) GoogleSreOptionFn {
	return func(option *googleSreOption) {
		option.windowTime = windowTime
	}
}

func WithGoogleSreBucket(count int) GoogleSreOptionFn {
	return func(option *googleSreOption) {
		option.bucketCount = count
	}
}

func NewGoogleSre(options ...GoogleSreOptionFn) *GoogleSre {
	op := defaultGoogleSreOption()
	for i := range options {
		options[i](op)
	}

	return &GoogleSre{
		option: op,
		rw:     rollingwindow.New(op.bucketCount, time.Duration(int64(op.windowTime)/int64(op.bucketCount))),
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func defaultGoogleSreOption() *googleSreOption {
	return &googleSreOption{
		windowTime:  time.Second * 10,
		bucketCount: 40,
		sreK:        1.5,
		protection:  5,
	}
}

func (g *GoogleSre) doReq(req func() error, reject func(err error) error, acceptable func(err error) bool) error {
	if err := g.accept(); err != nil {
		if reject != nil {
			return reject(err)
		}
		return err
	}

	defer func() {
		if e := recover(); e != nil {
			g.fail()
			panic(e)
		}
	}()

	err := req()
	if acceptable(err) {
		g.success()
	} else {
		g.fail()
	}
	return err
}

func (g *GoogleSre) accept() error {
	accepts, total := g.stat()
	weight := g.option.sreK * float64(accepts)
	// sreK++ ==> dropRatio--
	dropRatio := math.Max(0,
		float64(total)-(float64(g.option.protection)+weight)/float64(total+1),
	)

	if g.shouldDrop(dropRatio) {
		return ErrGoogleSreBreakOn
	}

	return nil
}

func (g *GoogleSre) shouldDrop(dropRatio float64) (should bool) {
	g.lock.Lock()
	defer g.lock.Unlock()
	if dropRatio <= 0 {
		return false
	}

	return g.rand.Float64() < dropRatio
}

func (g *GoogleSre) success() {
	g.rw.Add(1)
}

func (g *GoogleSre) fail() {
	// sum += 0, count++
	g.rw.Add(0)
}

func (g *GoogleSre) stat() (accepts, total int64) {
	g.rw.Reduce(func(bucket *rollingwindow.Bucket) {
		accepts += int64(bucket.Sum)
		total += bucket.Count
	})
	return accepts, total
}
