package ratelimit

import (
	"errors"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/sado0823/go-kitx/kit/log"
	"github.com/sado0823/go-kitx/pkg/atomicx"
	rollingwindow "github.com/sado0823/go-kitx/pkg/rollingwindow/v2"
	"github.com/sado0823/go-kitx/pkg/syncx"
	"github.com/sado0823/go-kitx/pkg/sysx"
	"github.com/sado0823/go-kitx/pkg/timex"
)

const (
	bbrBuckets         = 50
	bbrWindow          = time.Second * 5
	bbrCpuThreshold    = 900
	bbrMinRt           = float64(time.Second / time.Millisecond)
	bbrFlyingBeta      = 0.9
	bbrCoolingDuration = time.Second
)

var (
	ErrBBRServiceOverload = errors.New("service overload with bbr")

	overloadChecker = func(threshold int64) bool {
		return sysx.CpuUsage() >= threshold
	}
)

type (
	BBROptionFn func(*bbrOption)

	bbrOption struct {
		window       time.Duration
		buckets      int
		cpuThreshold int64
	}

	Promise interface {
		MarkSuccess()
		MarkFail()
	}

	promise struct {
		startTime time.Duration
		bbr       *BBR
	}

	// BBR Bottleneck Bandwidth and RTT
	BBR struct {
		option          *bbrOption
		windows         int64
		flying          int64
		ewmaFlying      float64
		ewmaFlyingLock  syncx.SpinLock
		dropTime        *atomicx.Duration
		droppedRecently *atomicx.Bool
		passCounter     *rollingwindow.RollingWindow
		rtCounter       *rollingwindow.RollingWindow
	}
)

func WithBBRWindow(window time.Duration) BBROptionFn {
	return func(option *bbrOption) {
		option.window = window
	}
}

func WithBBRBucket(buckets int) BBROptionFn {
	return func(option *bbrOption) {
		option.buckets = buckets
	}
}

func WithBBRCpuThreshold(threshold int64) BBROptionFn {
	return func(option *bbrOption) {
		option.cpuThreshold = threshold
	}
}

func NewBBR(options ...BBROptionFn) *BBR {
	op := &bbrOption{
		window:       bbrWindow,
		buckets:      bbrBuckets,
		cpuThreshold: bbrCpuThreshold,
	}

	for _, option := range options {
		option(op)
	}

	bucketDuration := op.window / time.Duration(op.buckets)

	return &BBR{
		option:          op,
		windows:         int64(time.Second / bucketDuration),
		ewmaFlyingLock:  syncx.SpinLock{},
		dropTime:        atomicx.NewAtomicDuration(),
		droppedRecently: atomicx.NewAtomicBool(),
		passCounter:     rollingwindow.New(op.buckets, bucketDuration, rollingwindow.WithIgnoreCurrent()),
		rtCounter:       rollingwindow.New(op.buckets, bucketDuration, rollingwindow.WithIgnoreCurrent()),
	}
}

func (b *promise) MarkSuccess() {
	b.bbr.addFlying(-1)

	rt := float64(timex.Since(b.startTime)) / float64(time.Millisecond)
	b.bbr.rtCounter.Add(math.Ceil(rt))
	b.bbr.passCounter.Add(1)

}

func (b *promise) MarkFail() {
	b.bbr.addFlying(-1)
}

func (b *BBR) Allow() (Promise, error) {
	if b.shouldDrop() {
		b.dropTime.Set(timex.Now())
		b.droppedRecently.Set(true)
		return nil, ErrBBRServiceOverload
	}

	// will sub this counter while this request is finished
	b.addFlying(1)

	return &promise{
		startTime: timex.Now(),
		bbr:       b,
	}, nil
}

func (b *BBR) addFlying(delta int64) {
	flying := atomic.AddInt64(&b.flying, delta)

	if delta < 0 {
		b.ewmaFlyingLock.Lock()
		b.ewmaFlying = b.ewmaFlying*bbrFlyingBeta + float64(flying)*(1-bbrFlyingBeta)
		b.ewmaFlyingLock.Unlock()
	}
}

func (b *BBR) shouldDrop() bool {
	if b.isOverloaded() || b.stillHot() {
		if b.highThru() {
			flying := atomic.LoadInt64(&b.flying)
			b.ewmaFlyingLock.Lock()
			avgFlying := b.ewmaFlying
			b.ewmaFlyingLock.Unlock()
			msg := fmt.Sprintf(
				"dropreq, cpu: %d, maxPass: %d, minRt: %.2f, hot: %t, flying: %d, ewmaFlying: %.2f, windows:%d",
				sysx.CpuUsage(), b.maxPass(), b.minRt(), b.stillHot(), flying, avgFlying, b.windows)
			log.Error(msg)
			return true
		}
	}
	return false
}

func (b *BBR) highThru() bool {
	b.ewmaFlyingLock.Lock()
	ewmaFlying := b.ewmaFlying
	b.ewmaFlyingLock.Unlock()
	maxFlight := b.maxFlight()

	return int64(ewmaFlying) > maxFlight && atomic.LoadInt64(&b.flying) > maxFlight
}

func (b *BBR) maxFlight() int64 {
	// windows = buckets per second
	// maxQPS = maxPASS * windows
	// minRT = min average response time in milliseconds
	// maxQPS * minRT / milliseconds_per_second
	return int64(math.Max(1, float64(b.maxPass()*b.windows)*(b.minRt()/1e3)))
}

func (b *BBR) maxPass() int64 {
	var res float64 = 1

	b.passCounter.Reduce(func(bucket *rollingwindow.Bucket) {
		if bucket.Sum > res {
			res = bucket.Sum
		}
	})
	return int64(res)
}

func (b *BBR) minRt() float64 {
	res := bbrMinRt

	b.rtCounter.Reduce(func(bucket *rollingwindow.Bucket) {
		if bucket.Count <= 0 {
			return
		}
		avg := math.Round(bucket.Sum / float64(bucket.Count))
		if avg < res {
			res = avg
		}
	})
	return res
}

func (b *BBR) stillHot() bool {
	if !b.droppedRecently.True() {
		return false
	}

	lastDropTime := b.dropTime.Load()
	if lastDropTime == 0 {
		return false
	}

	hot := timex.Since(lastDropTime) < bbrCoolingDuration
	if !hot {
		b.droppedRecently.Set(hot)
	}

	return hot
}

func (b *BBR) isOverloaded() bool {
	return overloadChecker(b.option.cpuThreshold)
}
