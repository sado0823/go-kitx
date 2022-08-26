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
	googleSreOptionFn func(*googleSreOption)

	googleSreOption struct {
		windowTime  time.Duration
		bucketCount int
		sreK        float64
		protection  int
	}

	googleSre struct {
		option *googleSreOption
		rw     *rollingwindow.RollingWindow
		lock   sync.Mutex
		rand   *rand.Rand
	}
)

func withGoogleSreWindow(windowTime time.Duration) googleSreOptionFn {
	return func(option *googleSreOption) {
		option.windowTime = windowTime
	}
}

func withGoogleSreBucket(count int) googleSreOptionFn {
	return func(option *googleSreOption) {
		option.bucketCount = count
	}
}

func newGoogleSre(options ...googleSreOptionFn) *googleSre {
	op := defaultGoogleSreOption()
	for i := range options {
		options[i](op)
	}

	return &googleSre{
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

func (g *googleSre) MarkSuccess() {
	g.rw.Add(1)
}

func (g *googleSre) MarkFail() {
	// sum += 0, count++
	g.rw.Add(0)
}

func (g *googleSre) Allow() error {
	accepts, total := g.stat()
	weight := g.option.sreK*float64(accepts) + float64(g.option.protection)
	// from 《google sre》
	// sreK++ ==> dropRatio--
	dropRatio := math.Max(0, (float64(total)-weight)/float64(total+1))

	if g.shouldDrop(dropRatio) {
		logger.Printf("accepts:%d, total:%d, dropRatio:%v", accepts, total, dropRatio)
		return ErrGoogleSreBreakOn
	}

	return nil
}

func (g *googleSre) stat() (accepts, total int64) {
	g.rw.Reduce(func(bucket *rollingwindow.Bucket) {
		accepts += int64(bucket.Sum)
		total += bucket.Count
	})
	return accepts, total
}

func (g *googleSre) shouldDrop(dropRatio float64) (should bool) {
	g.lock.Lock()
	defer g.lock.Unlock()
	if dropRatio <= 0 {
		return false
	}

	return g.rand.Float64() < dropRatio
}

func (g *googleSre) DoWithAcceptable(req func() error, acceptable func(err error) bool) error {
	return g.doReq(req, nil, acceptable)
}

func (g *googleSre) doReq(req func() error, reject func(err error) error, acceptable func(err error) bool) error {
	if err := g.Allow(); err != nil {
		if reject != nil {
			return reject(err)
		}
		return err
	}

	defer func() {
		if e := recover(); e != nil {
			g.MarkFail()
			panic(e)
		}
	}()

	err := req()
	if acceptable(err) {
		g.MarkSuccess()
	} else {
		g.MarkFail()
	}
	return err
}
