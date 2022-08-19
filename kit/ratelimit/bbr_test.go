package ratelimit

import (
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sado0823/go-kitx/pkg/atomicx"
	rollingwindow "github.com/sado0823/go-kitx/pkg/rollingwindow/v2"
	"github.com/sado0823/go-kitx/pkg/timex"
	"github.com/stretchr/testify/assert"
)

const (
	testBuckets        = 10
	testBucketDuration = time.Millisecond * 50
)

func newTestRW() *rollingwindow.RollingWindow {
	return rollingwindow.New(testBuckets, testBucketDuration, rollingwindow.WithIgnoreCurrent())
}

func Test_BBR_stillHot(t *testing.T) {
	passCounter := newTestRW()
	rtCounter := newTestRW()
	for i := 0; i < 10; i++ {
		if i > 0 {
			time.Sleep(testBucketDuration)
		}
		passCounter.Add(float64((i + 1) * 100))
		for j := i*10 + 1; j <= i*10+10; j++ {
			rtCounter.Add(float64(j))
		}
	}
	bbr := &BBR{
		option:          &bbrOption{cpuThreshold: 800},
		passCounter:     passCounter,
		rtCounter:       rtCounter,
		windows:         testBuckets,
		droppedRecently: atomicx.NewAtomicBool(),
		dropTime:        atomicx.NewAtomicDuration(),
	}
	t.Run("not dropped", func(t *testing.T) {
		assert.False(t, bbr.stillHot())
		bbr.dropTime.Set(-bbrCoolingDuration * 2)
		assert.False(t, bbr.stillHot())
	})

	t.Run("dropped but after cool duration", func(t *testing.T) {
		bbr.droppedRecently.Set(true)
		bbr.dropTime.Set(timex.Now() - bbrCoolingDuration*2)
		assert.False(t, bbr.stillHot())
	})

	t.Run("dropped but in cool duration", func(t *testing.T) {
		bbr.droppedRecently.Set(true)
		bbr.dropTime.Set(timex.Now() + bbrCoolingDuration/2)
		assert.True(t, bbr.stillHot())
	})

}

func Test_BBR_shouldDrop(t *testing.T) {
	passCounter := newTestRW()
	rtCounter := newTestRW()
	for i := 0; i < 10; i++ {
		if i > 0 {
			time.Sleep(testBucketDuration)
		}
		passCounter.Add(float64((i + 1) * 100))
		for j := i*10 + 1; j <= i*10+10; j++ {
			rtCounter.Add(float64(j))
		}
	}
	bbr := &BBR{
		option:          &bbrOption{cpuThreshold: 800},
		passCounter:     passCounter,
		rtCounter:       rtCounter,
		windows:         testBuckets,
		droppedRecently: atomicx.NewAtomicBool(),
		dropTime:        atomicx.NewAtomicDuration(),
	}

	// should drop = ewmaFlying > maxFlight && flying > maxFlight
	// maxFlight = 54
	t.Run("overload && ewmaFlying < maxFlight", func(t *testing.T) {
		overloadChecker = func(threshold int64) bool {
			return true
		}
		bbr.ewmaFlying = 50
		assert.False(t, bbr.shouldDrop())
	})

	t.Run("overload && ewmaFlying > maxFlight && flying < maxFlight", func(t *testing.T) {
		overloadChecker = func(threshold int64) bool {
			return true
		}
		bbr.ewmaFlying = 80
		bbr.flying = 50
		assert.False(t, bbr.shouldDrop())
	})

	t.Run("not overload && flying > maxFlight", func(t *testing.T) {
		overloadChecker = func(threshold int64) bool {
			return false
		}
		bbr.flying = 80
		assert.False(t, bbr.shouldDrop())
	})

	t.Run("overload && flying > maxFlight", func(t *testing.T) {
		overloadChecker = func(threshold int64) bool {
			return true
		}
		bbr.flying = 80
		bbr.ewmaFlying = 80
		_, err := bbr.Allow()
		assert.ErrorIs(t, err, ErrBBRServiceOverload)
	})

}

func Test_BBR_maxFlight(t *testing.T) {
	passCounter := newTestRW()
	rtCounter := newTestRW()
	for i := 0; i < 10; i++ {
		if i > 0 {
			time.Sleep(testBucketDuration)
		}
		passCounter.Add(float64((i + 1) * 100))
		for j := i*10 + 1; j <= i*10+10; j++ {
			rtCounter.Add(float64(j))
		}
	}
	bbr := &BBR{
		passCounter:     passCounter,
		rtCounter:       rtCounter,
		windows:         testBuckets,
		droppedRecently: atomicx.NewAtomicBool(),
	}
	assert.Equal(t, int64(54), bbr.maxFlight())
}

func Test_BBR_minRt(t *testing.T) {
	t.Run("default min rt ", func(t *testing.T) {
		rtCounter := newTestRW()
		bbr := BBR{
			droppedRecently: atomicx.NewAtomicBool(),
			rtCounter:       rtCounter,
		}
		assert.Equal(t, bbrMinRt, bbr.minRt())
	})

	t.Run("normal", func(t *testing.T) {
		rtCounter := newTestRW()
		for i := 0; i < 10; i++ {
			if i > 0 {
				time.Sleep(testBucketDuration)
			} else {
				for j := i*10 + 1; j <= i*10+10; j++ {
					rtCounter.Add(float64(j))
				}
			}
		}
		bbr := BBR{
			droppedRecently: atomicx.NewAtomicBool(),
			rtCounter:       rtCounter,
		}
		// expect 6
		assert.Equal(t, math.Round(float64(1+2+3+4+5+6+7+8+9+10)/10), bbr.minRt())
	})
}

func Test_BBR_maxPass(t *testing.T) {
	t.Run("default max pass = 1", func(t *testing.T) {
		passCounter := newTestRW()
		bbr := BBR{
			droppedRecently: atomicx.NewAtomicBool(),
			passCounter:     passCounter,
		}
		assert.Equal(t, int64(1), bbr.maxPass())
	})

	t.Run("normal", func(t *testing.T) {
		passCounter := newTestRW()
		for i := 0; i <= 10; i++ {
			passCounter.Add(float64(i * 100))
			time.Sleep(testBucketDuration)
		}

		bbr := BBR{
			droppedRecently: atomicx.NewAtomicBool(),
			passCounter:     passCounter,
		}
		assert.Equal(t, int64(1000), bbr.maxPass())
	})
}

func Test_BBR(t *testing.T) {
	bbr := NewBBR(WithBBRWindow(testBucketDuration), WithBBRBucket(testBuckets), WithBBRCpuThreshold(100))

	var (
		probaRand = rand.New(rand.NewSource(time.Now().UnixNano()))
		probaLock sync.Mutex
		wg        sync.WaitGroup
		drop      int64
	)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 30; i++ {
				promise, err := bbr.Allow()
				if err != nil {
					atomic.AddInt64(&drop, 1)
				} else {
					count := rand.Intn(5)
					time.Sleep(time.Millisecond * time.Duration(count))
					probaLock.Lock()
					fail := probaRand.Float64() < 0.01
					probaLock.Unlock()
					if fail {
						promise.MarkFail()
					} else {
						promise.MarkSuccess()
					}
				}
			}
		}()
	}
	wg.Wait()
}

func Benchmark_BBR(b *testing.B) {
	bench := func(b *testing.B) {
		shedder := NewBBR()
		var (
			probaRand = rand.New(rand.NewSource(time.Now().UnixNano()))
			probaLock sync.Mutex
		)
		for i := 0; i < 6000; i++ {
			promise, err := shedder.Allow()
			if err == nil {
				time.Sleep(time.Millisecond)
				probaLock.Lock()
				fail := probaRand.Float64() < 0.01
				probaLock.Unlock()
				if fail {
					promise.MarkFail()
				} else {
					promise.MarkSuccess()
				}
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			promise, err := shedder.Allow()
			if err == nil {
				promise.MarkSuccess()
			}
		}
	}

	b.Run("high load", func(b *testing.B) {
		overloadChecker = func(int64) bool {
			return true
		}
		bench(b)
	})

	b.Run("low load", func(b *testing.B) {
		overloadChecker = func(int64) bool {
			return false
		}
		bench(b)
	})
}
