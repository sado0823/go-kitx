package internal

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testStep = time.Minute
	waitTime = time.Second
)

func Test_NewTimingWheel(t *testing.T) {

	testCases := []struct {
		name     string
		interval time.Duration
		numSlots int
		execute  func(key, value interface{})
		pass     bool
	}{
		{
			// error interval
			name: "error interval", interval: 0, numSlots: 10, execute: nil, pass: false,
		},
		{
			// error numSlots
			name: "error numSlots", interval: testStep, numSlots: 0, execute: nil, pass: false,
		},
		{
			// err execute
			name: "err execute", interval: testStep, numSlots: 10, execute: nil, pass: false,
		},
		{
			// correct
			name: "correct", interval: testStep, numSlots: 10, execute: func(key, value interface{}) {}, pass: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tw, err := NewTimingWheel(testCase.interval, testCase.numSlots, testCase.execute)
			if testCase.pass {
				assert.Nil(t, err)
				defer tw.Stop()
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

// 立刻执行全部任务
func Test_TimingWheel_Drain(t *testing.T) {
	ticker := NewFakeTicker()
	tw, _ := newTimingWheel(testStep, 10, func(key, value interface{}) {}, ticker)
	defer tw.Stop()

	tw.SetTimer("a", 1, testStep*1)
	tw.SetTimer("b", 3, testStep*3)
	tw.SetTimer("c", 5, testStep*5)

	var (
		keys []string
		vals []int
		mu   sync.Mutex
		wg   sync.WaitGroup
	)

	wg.Add(3)
	tw.Drain(func(key, value interface{}) {
		mu.Lock()
		defer mu.Unlock()
		defer wg.Done()
		keys = append(keys, key.(string))
		vals = append(vals, value.(int))
	})

	wg.Wait()

	assert.ElementsMatch(t, keys, []string{"a", "b", "c"})
	assert.ElementsMatch(t, vals, []int{1, 3, 5})

	var counter int
	tw.Drain(func(key, value interface{}) {
		counter++
	})
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, 0, counter)
}

// 任务时间小于轮盘时钟滚动时间, 会立即执行任务
func Test_TimingWheel_SetTimerSoon(t *testing.T) {
	run := NewAtomicBool()
	ticker := NewFakeTicker()

	tw, _ := newTimingWheel(testStep, 10, func(key, value interface{}) {
		assert.True(t, run.CompareAndSwap(false, true))
		assert.Equal(t, "any", key)
		assert.Equal(t, 3, value.(int))
		ticker.Done()
	}, ticker)
	defer tw.Stop()

	tw.SetTimer("any", 3, testStep>>1)
	ticker.Tick()
	assert.Nil(t, ticker.Wait(waitTime))
	assert.True(t, run.True())
}

// 统一任务设置不同的执行时间, 以较长的时间为准
func Test_TimingWheel_SetTimerTwice(t *testing.T) {
	run := NewAtomicBool()
	ticker := NewFakeTicker()

	tw, _ := newTimingWheel(testStep, 10, func(key, value interface{}) {
		assert.True(t, run.CompareAndSwap(false, true))
		assert.Equal(t, "a", key)
		assert.Equal(t, 5, value.(int))
		ticker.Done()
	}, ticker)
	defer tw.Stop()

	tw.SetTimer("a", 3, testStep*3)
	tw.SetTimer("a", 5, testStep*5)

	for i := 0; i < 6; i++ {
		ticker.Tick()
	}

	assert.Nil(t, ticker.Wait(waitTime))
	assert.True(t, run.True())
}

func Test_TimingWheel_SetTimerWrongDelay(t *testing.T) {
	ticker := NewFakeTicker()

	tw, _ := newTimingWheel(testStep, 10, func(key, value interface{}) {}, ticker)
	defer tw.Stop()

	assert.NotPanics(t, func() {
		tw.SetTimer("a", 3, -testStep)
	})
}

// 移动时间轮任务
func Test_TimingWheel_MoveTimer(t *testing.T) {
	run := NewAtomicBool()
	ticker := NewFakeTicker()

	tw, _ := newTimingWheel(testStep, 3, func(key, value interface{}) {
		assert.True(t, run.CompareAndSwap(false, true))
		assert.Equal(t, "a", key)
		assert.Equal(t, 3, value.(int))
		ticker.Done()
	}, ticker)

	tw.SetTimer("a", 3, testStep*4)
	tw.MoveTimer("a", testStep*7)
	tw.MoveTimer("a", -testStep*7)
	tw.MoveTimer("none", testStep)

	for i := 0; i < 5; i++ {
		ticker.Tick()
	}
	assert.False(t, run.True())

	for i := 0; i < 3; i++ {
		ticker.Tick()
	}
	assert.Nil(t, ticker.Wait(waitTime))
	assert.True(t, run.True())
}

// 移动时间轮, 并且移动的时间小于时间轮刻度
func Test_TimingWheel_MoveTimerSoon(t *testing.T) {
	run := NewAtomicBool()
	ticker := NewFakeTicker()

	tw, _ := newTimingWheel(testStep, 3, func(key, value interface{}) {
		assert.True(t, run.CompareAndSwap(false, true))
		assert.Equal(t, "a", key)
		assert.Equal(t, 3, value.(int))
		ticker.Done()
	}, ticker)
	defer tw.Stop()

	tw.SetTimer("a", 3, testStep*4)
	tw.MoveTimer("a", testStep>>1)

	assert.Nil(t, ticker.Wait(waitTime))
	assert.True(t, run.True())
}

func Test_TimingWheel_MoveTimerEarlier(t *testing.T) {
	run := NewAtomicBool()
	ticker := NewFakeTicker()

	tw, _ := newTimingWheel(testStep, 3, func(key, value interface{}) {
		assert.True(t, run.CompareAndSwap(false, true))
		assert.Equal(t, "a", key)
		assert.Equal(t, 3, value.(int))
		ticker.Done()
	}, ticker)
	defer tw.Stop()

	tw.SetTimer("a", 3, testStep*7)
	tw.MoveTimer("a", testStep*3)

	for i := 0; i < 4; i++ {
		ticker.Tick()
	}

	assert.Nil(t, ticker.Wait(waitTime))
	assert.True(t, run.True())
}

func Test_TimingWheel_RemoveTimer(t *testing.T) {
	run := NewAtomicBool()
	ticker := NewFakeTicker()

	tw, _ := newTimingWheel(testStep, 10, func(key, value interface{}) {
		run.CompareAndSwap(false, true)
		ticker.Done()
	}, ticker)
	tw.SetTimer("a", 3, testStep*3)

	assert.NotPanics(t, func() {
		tw.RemoveTimer("a")
		tw.RemoveTimer("none")
		tw.RemoveTimer(nil)
	})

	for i := 0; i < 4; i++ {
		ticker.Tick()
	}

	tw.Stop()
	assert.False(t, run.True())
}

func Test_TimingWheel_MoveAndRemoveTimer(t *testing.T) {
	ticker := NewFakeTicker()
	tick := func(counter int) {
		for i := 0; i < counter; i++ {
			ticker.Tick()
		}
	}

	var keys []int
	tw, _ := newTimingWheel(testStep, 10, func(key, value interface{}) {
		assert.Equal(t, "foo", key)
		assert.Equal(t, 3, value.(int))
		keys = append(keys, value.(int))
		ticker.Done()
	}, ticker)
	defer tw.Stop()

	tw.SetTimer("foo", 3, testStep*8)
	tick(6)

	tw.MoveTimer("foo", testStep*7)
	tick(3)
	assert.Equal(t, 0, len(keys))

	tw.RemoveTimer("foo")
	tick(30)
	time.Sleep(time.Millisecond)
	assert.Equal(t, 0, len(keys))
}

func BenchmarkTimingWheel(b *testing.B) {
	b.ReportAllocs()

	tw, _ := NewTimingWheel(time.Second, 100, func(k, v interface{}) {})
	for i := 0; i < b.N; i++ {
		tw.SetTimer(i, i, time.Second)
		tw.SetTimer(b.N+i, b.N+i, time.Second)
		tw.MoveTimer(i, time.Second*time.Duration(i))
		tw.RemoveTimer(i)
	}
}
