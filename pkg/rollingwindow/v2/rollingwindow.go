package v2

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

var (
	logger = log.New(os.Stdout, fmt.Sprintf("[DEBUG][pkg=rollingwindow/v2][%s] ", time.Now().Format(time.StampMilli)), log.Lshortfile)
)

func init() {
	logger.SetFlags(0)
	logger.SetOutput(io.Discard)
}

const unixTimeUnitOffset = int64(time.Millisecond / time.Nanosecond)

type (
	OptionFn func(*Option)

	Option struct {
		ignoreCurrent bool
	}

	ReduceFn func(*Bucket)

	RollingWindow struct {
		rw             sync.RWMutex
		size           int
		eachBucketTime time.Duration
		window         *window
		offset         int
		lastTime       time.Time

		option *Option
	}
)

func WithIgnoreCurrent() OptionFn {
	return func(option *Option) {
		option.ignoreCurrent = true
	}
}

// New 计算上次当前时间与上次时间差的窗口周期跨度(采用跨度span为标准), 返回跨度为1两侧的数据
func New(size int, eachBucketTime time.Duration, options ...OptionFn) *RollingWindow {
	if size <= 0 {
		panic("size must greater than 0")
	}

	op := defaultOption()
	for i := range options {
		options[i](op)
	}

	return &RollingWindow{
		rw:             sync.RWMutex{},
		size:           size,
		eachBucketTime: eachBucketTime,
		window:         newWindow(size),
		lastTime:       time.Now(),
		option:         op,
	}
}

func (r *RollingWindow) Add(v float64) {
	r.rw.Lock()
	defer r.rw.Unlock()

	now := time.Now()
	r.adjust(now)
	r.window.add(r.offset, v)
}

func (r *RollingWindow) Reduce(fn ReduceFn) {
	r.rw.RLock()
	defer r.rw.RUnlock()

	now := time.Now()
	span := r.span(now)

	var diff int
	if span == 0 && r.option.ignoreCurrent {
		diff = r.size - 1
	} else {
		diff = r.size - span
	}
	if diff > 0 {
		r.window.reduce(r.offset+span+1, diff, fn)
	}
}

func (r *RollingWindow) adjust(now time.Time) {
	span := r.span(now)
	if span <= 0 {
		return
	}

	oldOffset := r.offset
	// reset expired buckets
	for i := 0; i < span; i++ {
		logger.Println("adjust reset offset: ", oldOffset+i+1, "sum: ", r.window.find(oldOffset+i+1).Sum)
		r.window.resetOne(oldOffset + i + 1)
	}

	newOffset := (oldOffset + span) % r.size
	logger.Println("adjust oldOffset: ", oldOffset, "newOffset: ", newOffset, "span: ", span, "diff: ", r.diffMilli(now))

	nowMilli := r.milliByTime(now)
	lastTimeMilli := r.milliByTime(r.lastTime)
	adjustTime := nowMilli - (nowMilli-lastTimeMilli)%r.milliByDuration(r.eachBucketTime)

	r.offset = newOffset
	r.lastTime = time.UnixMilli(adjustTime)
}

func (r *RollingWindow) span(now time.Time) (span int) {
	spanV := r.diffMilli(now) / r.milliByDuration(r.eachBucketTime)
	if spanV >= 0 && spanV < int64(r.size) {
		return int(spanV)
	}

	// span == r.size, enough to clear expired bucket
	return r.size
}

func (r *RollingWindow) diffMilli(now time.Time) int64 {
	return r.milliByDuration(now.Sub(r.lastTime))
}

func (r *RollingWindow) milliByTime(now time.Time) int64 {
	return now.UnixNano() / unixTimeUnitOffset
}

func (r *RollingWindow) milliByDuration(duration time.Duration) int64 {
	return duration.Nanoseconds() / unixTimeUnitOffset
}

func defaultOption() *Option {
	return &Option{}
}

func newWindow(size int) *window {
	buckets := make([]*Bucket, size)
	for i := range buckets {
		buckets[i] = new(Bucket)
	}
	return &window{
		buckets: buckets,
		size:    size,
	}
}

type window struct {
	buckets []*Bucket
	size    int
}

func (w *window) find(offset int) *Bucket {
	return w.buckets[offset%w.size]
}

func (w *window) add(offset int, v float64) {
	w.find(offset).add(v)
}

func (w *window) reduce(start, count int, fn ReduceFn) {
	for i := 0; i < count; i++ {
		fn(w.find(start + i))
	}
}

func (w *window) resetOne(offset int) {
	w.find(offset).reset()
}

type Bucket struct {
	Sum   float64
	Count int64
}

func (b *Bucket) add(v float64) {
	b.Sum += v
	b.Count++
}

func (b *Bucket) reset() {
	b.Sum = 0
	b.Count = 0
}
