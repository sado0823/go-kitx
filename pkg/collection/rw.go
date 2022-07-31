package collection

import (
	"fmt"
	"sync"
	"time"
)

const unixTimeUnitOffset = int64(time.Millisecond / time.Nanosecond)

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
		//lastTime time.Time

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
		//lastTime: time.Now(),
		option: op,
	}
}

func (r *RollingWindow) Add(v float64) {
	r.rw.Lock()
	defer r.rw.Unlock()
	now := time.Now()
	bucket := r.currentBucket(now)
	bucket.add(v)
}

func (r *RollingWindow) Reduce(fn ReduceFn) {
	r.rw.RLock()
	defer r.rw.RUnlock()

	now := time.Now()
	//round := r.round(now)
	start, end := r.timeRange(now)

	for i := range r.window.buckets {
		b := r.window.buckets[i]
		fmt.Printf("reduce bucket: %#v, index:%d \n", b, i)
		if b != nil && b.startTime >= start && b.startTime <= end {
			fn(b)
			continue
		}
		fn(&Bucket{})
	}
}

func (r *RollingWindow) currentBucket(now time.Time) *Bucket {
	offset := r.index(now)
	//round := r.round(now)
	startTime := r.startTime(now)
	bucket := r.window.find(offset)

	fmt.Printf("currentBucket: %#v \n", bucket)
	if bucket == nil {
		r.window.buckets[offset] = &Bucket{
			Sum:       0,
			Count:     0,
			startTime: startTime,
		}
		return r.window.buckets[offset]
	} else if startTime == bucket.startTime {
		return bucket
	} else if startTime > bucket.startTime {
		bucket.Sum = 0
		bucket.Count = 0
		bucket.startTime = startTime
		return bucket
	} else {
		panic(fmt.Sprintf("invalid bucket: %d", now.UnixMilli()))
	}
}

func (r *RollingWindow) index(now time.Time) (offset int) {
	eachBucket := r.milliByDuration(r.interval) / int64(r.size)
	index := int(r.milliByTime(now) / eachBucket % int64(r.size))
	fmt.Println("index:", index)
	return index
	// ====
	//diff := r.milliByDuration(now.Sub(r.lastTime))
	//eachBucket := r.milliByDuration(r.interval) / int64(r.size)
	//offsetV := (diff / eachBucket) % int64(r.size)
	//fmt.Println("index diff: ", diff, "offsetV: ", offsetV)
	//return int(offsetV)
}

func (r *RollingWindow) timeRange(now time.Time) (start, end int64) {
	end = r.startTime(now)
	eachBucket := r.milliByDuration(r.interval) / int64(r.size)
	start = end - (r.milliByDuration(r.interval) + eachBucket)
	return start, end
}

func (r *RollingWindow) startTime(now time.Time) int64 {
	eachBucket := r.milliByDuration(r.interval) / int64(r.size)
	return r.milliByTime(now) - r.milliByTime(now)%eachBucket

	// ====
	//diff := r.milliByDuration(now.Sub(r.lastTime))
	//startTime := r.milliByTime(now) - diff%r.milliByDuration(r.interval)
	//return startTime
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
	startTime int64
	Sum       float64
	Count     int
}

func (b *Bucket) add(v float64) {
	b.Sum += v
	b.Count++
}

func (b *Bucket) reset() {
	b.Sum = 0
	b.Count = 0
}
