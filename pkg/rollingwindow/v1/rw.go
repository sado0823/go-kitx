package v1

import (
	"fmt"
	"math"
	"sync"
	"time"
)

const unixTimeUnitOffset = int64(time.Millisecond / time.Nanosecond)

type (
	OptionFn func(*Option)

	Option struct{}

	ReduceFn func(*Bucket)

	RollingWindow struct {
		rw              sync.RWMutex
		size            int
		interval        time.Duration
		window          *window
		initTime        time.Time
		eachBucketMilli int64

		option *Option
	}
)

// New 计算当前时间所属window的开始时间, 以此时间为截止时间(采用时间bucket.startTime当标准), 返回之前一个窗口周期内的bucket
func New(size int, interval time.Duration, options ...OptionFn) *RollingWindow {
	if size <= 0 {
		panic("size must greater than 0")
	}

	op := defaultOption()
	for i := range options {
		options[i](op)
	}

	r := &RollingWindow{
		rw:       sync.RWMutex{},
		size:     size,
		interval: interval,
		window:   newWindow(size),
		initTime: time.Now(),
		option:   op,
	}
	r.eachBucketMilli = r.milliByDuration(r.interval) / int64(r.size)

	return r
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
	start, end := r.timeRange(now)
	empty := &Bucket{}
	for i := range r.window.buckets {
		b := r.window.buckets[i]
		fmt.Printf("reduce bucket: %#v, index:%d start:%d, end:%d \n", b, i, start, end)
		if b != nil && b.startTime >= start && (b.startTime <= end || math.Abs(float64(b.startTime-end)) <= float64(r.eachBucketMilli/int64(r.size))) {
			fn(b)
			continue
		}
		fn(empty)
	}
}

func (r *RollingWindow) currentBucket(now time.Time) *Bucket {
	offset := r.index(now)
	startTime := r.startTime(now)
	bucket := r.window.find(offset)

	fmt.Printf("currentBucket: %#v \n", bucket)
	if bucket == nil {
		r.window.buckets[offset] = r.window.newBucket(startTime)
		return r.window.buckets[offset]
	} else if math.Abs(float64(startTime-bucket.startTime)) <= float64(r.eachBucketMilli/int64(r.size)) {
		fmt.Println("currentBucket startTime diff: ", startTime-bucket.startTime, "allow span: ", r.eachBucketMilli/int64(r.size))
		return bucket
	} else if startTime > bucket.startTime {
		bucket.reset(startTime)
		return bucket
	} else {
		panic(fmt.Sprintf("invalid bucket, now:%d, bucket:%#v, startTime:%d", now.UnixMilli(), bucket, startTime))
	}
}

func (r *RollingWindow) index(now time.Time) (offset int) {
	diff := r.diffMilli(now)
	offsetV := (diff / r.eachBucketMilli) % int64(r.size)
	fmt.Println("index diff: ", diff, "offsetV: ", offsetV)
	return int(offsetV)
}

func (r *RollingWindow) timeRange(now time.Time) (start, end int64) {
	end = r.startTime(now)
	start = end - (r.milliByDuration(r.interval) + r.eachBucketMilli)
	return start, end
}

func (r *RollingWindow) startTime(now time.Time) int64 {
	diff := r.diffMilli(now)
	startTime := r.milliByTime(now) - diff%r.milliByDuration(r.interval)
	return startTime
}

func (r *RollingWindow) diffMilli(now time.Time) int64 {
	return r.milliByDuration(now.Sub(r.initTime))
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

func (w *window) newBucket(startTime int64) *Bucket {
	return &Bucket{startTime: startTime}
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

func (b *Bucket) reset(startTime int64) {
	b.startTime = startTime
	b.Sum = 0
	b.Count = 0
}
