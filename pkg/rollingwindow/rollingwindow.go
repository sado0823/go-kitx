package rollingwindow

import (
	"sync"
	"time"
)

type (
	OptionFn func(*Option)

	Option struct {
	}

	ReduceFn func(*Bucket)

	RollingWindow struct {
		rw       sync.RWMutex
		size     int
		interval time.Duration
		window   *window
		offset   int
		lastTime time.Time

		option *Option
	}
)

func New(size int, interval time.Duration, options ...OptionFn) *RollingWindow {
	if size <= 0 {
		panic("size must greater than 0")
	}

	op := defaultOption()
	for i := range options {
		options[i](op)
	}

	return &RollingWindow{
		rw:       sync.RWMutex{},
		size:     size,
		interval: interval,
		window:   newWindow(size),
		lastTime: time.Now(),
		option:   op,
	}
}

func (r *RollingWindow) Add(v float64) {
	r.rw.Lock()
	defer r.rw.Unlock()

	now := time.Now()
	span := r.selectBucket(now)
	adjustOffset, adjustTime := r.adjust(span, now)

	r.offset = adjustOffset
	r.lastTime = adjustTime

	r.window.add(adjustOffset, v)
}

func (r *RollingWindow) Reduce(fn ReduceFn) {
	r.rw.RLock()
	defer r.rw.RUnlock()

	now := time.Now()
	span := r.selectBucket(now)
	adjustOffset, _ := r.adjust(span, now)
	r.window.reduce(adjustOffset, r.size, fn)
}

func (r *RollingWindow) adjust(span int, now time.Time) (adjustOffset int, adjustTime time.Time) {
	if span <= 0 {
		return r.offset, r.lastTime
	}

	offset := r.offset
	// reset expired buckets
	for i := 0; i < span; i++ {
		r.window.resetOne(offset + i + 1)
	}

	adjustOffset = (offset + span) % r.size

	diff := now.UnixMilli() - now.Sub(r.lastTime).Milliseconds()%r.interval.Milliseconds()
	adjustTime = time.UnixMilli(diff)

	return adjustOffset, adjustTime
}

func (r *RollingWindow) selectBucket(now time.Time) (span int) {
	diff := now.Sub(r.lastTime).Milliseconds()
	spanV := diff / r.interval.Milliseconds()
	//eachBucket := r.interval.Milliseconds() / int64(r.size)
	//offsetV := (diff / eachBucket) % int64(r.size)
	if span >= 0 && span < r.size {
		return int(spanV)
	}

	return r.size
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
		fn(w.find(start + 1 + i))
	}
}

func (w *window) resetOne(offset int) {
	w.find(offset).reset()
}

type Bucket struct {
	Sum   float64
	Count int
}

func (b *Bucket) add(v float64) {
	b.Sum += v
	b.Count++
}

func (b *Bucket) reset() {
	b.Sum = 0
	b.Count = 0
}
