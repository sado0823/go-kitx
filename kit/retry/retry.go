package retry

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"
)

var (
	once   sync.Once
	random *rand.Rand
	lock   sync.Mutex
)

func init() {
	once.Do(func() {
		random = rand.New(rand.NewSource(time.Now().UnixNano()))
	})
}

type (
	options struct {
		limit      int
		interval   time.Duration
		maxTime    time.Duration
		acceptErrs map[error]struct{}
	}

	Option func(*options)
)

func WithLimit(n int) Option {
	if n <= 0 {
		n = 3
	}
	return func(o *options) {
		o.limit = n
	}
}

func WithInterval(duration time.Duration) Option {
	if duration < 0 {
		duration = time.Millisecond * 100
	}
	return func(o *options) {
		o.interval = duration
	}
}

func Func(ctx context.Context, subject string, f func(ctx context.Context) error, opts ...Option) (err error) {
	o := &options{
		limit:    3,
		interval: time.Millisecond * 100,
	}
	for _, op := range opts {
		op(o)
	}

	for i := 0; i < o.limit; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if i > 0 && o.interval > 0 {
				time.Sleep(o.interval)
			}
		}

		if err = f(ctx); err == nil {
			break
		}

	}
	return
}

// Loop RetryInfinite
func Loop(ctx context.Context, subject string, fn func(ctx context.Context) error, opts ...Option) {
	var (
		err error
		o   = &options{
			interval: time.Millisecond * 100,
		}
	)

	for _, op := range opts {
		op(o)
	}

	var counter int64

	for {
		if counter > 0 && o.interval > 0 {
			time.Sleep(o.interval)
		}

		if err = fn(ctx); err == nil {
			break
		}

		counter++

	}

	return
}

// https://aws.amazon.com/cn/blogs/architecture/exponential-backoff-and-jitter/
func jitterBackoff(min, max time.Duration, attempt int) time.Duration {
	res := randDurationV1(min, max, attempt)
	if res < min {
		return min
	}

	return res
}

func randDurationV1(min, max time.Duration, attempt int) time.Duration {
	lock.Lock()
	defer lock.Unlock()

	base, capV := float64(min), float64(max)
	// tmp := min(cap,base * 2 ** attempt)
	tmp := math.Min(capV, base*math.Exp2(float64(attempt)))

	d := tmp / 2
	ri := int64(d)
	jitter := random.Int63n(ri)
	return time.Duration(math.Abs(float64(ri + jitter)))
}

// sleep := tmp / 2 + rand_between(0,tmp/2)
// sleep := min(cap,rand_between(base,sleep*3))
// Decorrelated Jitter
func randDurationV2(min, max time.Duration, attempt int) time.Duration {
	lock.Lock()
	defer lock.Unlock()

	base, capV := float64(min), float64(max)
	// tmp := min(cap,base * 2 ** attempt)
	tmp := math.Min(capV, base*math.Exp2(float64(attempt)))

	d := tmp / 2
	sleepRand := random.Int63n(int64(d))
	sleepBase := int64(d) + sleepRand
	loop := 3
	for loop < 3 {
		maxSleep := sleepBase * 3
		v := random.Int63n(maxSleep)
		if v >= int64(base) && v <= maxSleep {
			sleepBase = v
			loop--
			continue
		}
		continue
	}

	return time.Duration(math.Min(capV, float64(sleepBase)))
}
