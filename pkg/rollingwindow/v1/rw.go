package v1

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"sync"
	"time"
)

var (
	logger = log.New(os.Stdout, fmt.Sprintf("[DEBUG][pkg=rollingwindow/v1][%s] ", time.Now().Format(time.StampMilli)), log.Lshortfile)
)

func init() {
	logger.SetFlags(0)
	logger.SetOutput(io.Discard)
}

const unixTimeUnitOffset = int64(time.Millisecond / time.Nanosecond)

type (
	ReduceFn func(*Bucket)

	RollingWindow struct {
		rw             sync.RWMutex
		size           int
		eachBucketTime time.Duration
		window         *window
		initTime       time.Time

		eachBucketMilli int64
		windowMilli     int64
		withinMilli     float64 // 允许的误差值
	}
)

// New 计算当前时间所属window的开始时间, 以此时间为截止时间(采用时间bucket.startTime当标准), 返回之前一个窗口周期内的bucket
func New(size int, eachBucketTime time.Duration) *RollingWindow {
	if size <= 0 {
		panic("size must greater than 0")
	}

	r := &RollingWindow{
		rw:             sync.RWMutex{},
		size:           size,
		eachBucketTime: eachBucketTime,
		window:         newWindow(size),
		initTime:       time.Now(),
	}
	r.eachBucketMilli = r.milliByDuration(r.eachBucketTime)
	r.windowMilli = r.eachBucketMilli * int64(r.size)
	r.withinMilli = r.within()
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
		logger.Printf("reduce bucket: %#v, index:%d start:%d, end:%d , now:%d\n", b, i, start, end, r.milliByTime(now))
		if b != nil && b.bucketStartTime >= start && b.bucketStartTime <= end {
			fn(b)
			continue
		}

		//if b != nil && (r.allow(b.windowStartTime, end) || r.allow(b.windowStartTime, start)) {
		//	fn(b)
		//	continue
		//}
		fn(empty)
	}
}

func (r *RollingWindow) currentBucket(now time.Time) *Bucket {
	offset := r.index(now)
	startTime := r.windowStartTime(now)
	bucket := r.window.find(offset)
	bucketStartTime := startTime + int64(offset)*r.eachBucketMilli

	logger.Printf("currentBucket: %#v, windowStartTime:%d \n", bucket, startTime)
	if bucket == nil {
		r.window.buckets[offset] = r.window.newBucket(startTime, bucketStartTime)
		return r.window.buckets[offset]
		//} else if r.allow(windowStartTime, bucket.windowStartTime) {
	} else if r.allow(startTime, bucket.windowStartTime) {
		return bucket
	} else if startTime > bucket.windowStartTime {
		bucket.reset(startTime, bucketStartTime)
		return bucket
	} else {
		panic(fmt.Sprintf("invalid bucket, now:%d, bucket:%#v, windowStartTime:%d, r.eachBucketMilli:%d, r.size:%v", now.UnixMilli(), bucket, startTime, r.eachBucketMilli, r.size))
	}
}

func (r *RollingWindow) index(now time.Time) (offset int) {
	diff := r.diffMilli(now)
	offsetV := (diff / r.eachBucketMilli) % int64(r.size)
	logger.Println("index diff: ", diff, "offsetV: ", offsetV)
	return int(offsetV)
}

func (r *RollingWindow) timeRange(now time.Time) (start, end int64) {
	offset := r.index(now)
	end = r.windowStartTime(now) + int64(offset+1)*r.eachBucketMilli
	start = end - r.windowMilli
	return start - int64(r.withinMilli), end + int64(r.withinMilli)
}

func (r *RollingWindow) windowStartTime(now time.Time) int64 {
	diff := r.diffMilli(now)
	startTime := r.milliByTime(now) - diff%r.windowMilli
	return startTime
}

func (r *RollingWindow) allow(milli int64, compare int64) bool {
	actual := math.Abs(float64(milli - compare))
	logger.Printf("allow:%v,actual:%v, milli:%v,compare:%v \n", r.withinMilli, actual, milli, compare)
	return actual <= r.withinMilli
}

// 允许的误差范围
func (r *RollingWindow) within() float64 {
	allow := float64(r.eachBucketMilli / int64(r.size*3))
	if allow <= 0 {
		allow = 5
	}
	return allow
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

func (w *window) newBucket(windowStartTime, bucketStartTime int64) *Bucket {
	return &Bucket{windowStartTime: windowStartTime, bucketStartTime: bucketStartTime}
}

type Bucket struct {
	windowStartTime int64
	bucketStartTime int64
	Sum             float64
	Count           int64
}

func (b *Bucket) add(v float64) {
	b.Sum += v
	b.Count++
}

func (b *Bucket) reset(windowStartTime, bucketStartTime int64) {
	b.windowStartTime = windowStartTime
	b.bucketStartTime = bucketStartTime
	b.Sum = 0
	b.Count = 0
}
