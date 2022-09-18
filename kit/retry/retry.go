package retry

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"
)

const (
	defaultMaxRetries  = 3
	defaultMinWaitTime = time.Millisecond * 100
	defaultMaxWaitTime = time.Millisecond * 2000
)

var (
	once   sync.Once
	random *rand.Rand
)

func init() {
	once.Do(func() {
		random = rand.New(rand.NewSource(time.Now().UnixNano()))
	})
}

type (
	options struct {
		limit       int
		minWaitTime time.Duration
		maxWaitTime time.Duration
	}

	Option func(*options)
)

func WithLimit(n int) Option {
	if n <= 0 {
		n = defaultMaxRetries
	}
	return func(o *options) {
		o.limit = n
	}
}

func WithMin(min time.Duration) Option {
	if min <= 0 {
		min = defaultMinWaitTime
	}
	return func(o *options) {
		o.minWaitTime = min
	}
}

func WithMax(max time.Duration) Option {
	if max <= 0 {
		max = defaultMaxWaitTime
	}
	return func(o *options) {
		o.maxWaitTime = max
	}
}

func Func(ctx context.Context, subject string, fn func(ctx context.Context) error, opts ...Option) (err error) {
	o := &options{
		limit:       defaultMaxRetries,
		minWaitTime: defaultMinWaitTime,
		maxWaitTime: defaultMaxWaitTime,
	}
	for _, op := range opts {
		op(o)
	}

	for attempt := 0; attempt <= o.limit; attempt++ {
		err = fn(ctx)
		if err == nil {
			break
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		backoff := jitterBackoff(o.minWaitTime, o.maxWaitTime, attempt)

		t := time.NewTimer(backoff)
		select {
		case <-t.C:
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		}
	}

	return err
}

// Loop RetryInfinite
func Loop(ctx context.Context, subject string, fn func(ctx context.Context) error, opts ...Option) error {
	var (
		err error
		o   = &options{
			minWaitTime: defaultMinWaitTime,
			maxWaitTime: defaultMaxWaitTime,
		}
	)

	for _, op := range opts {
		op(o)
	}

	var counter int
	for {
		err = fn(ctx)
		if err == nil {
			break
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		backoff := jitterBackoff(o.minWaitTime, o.maxWaitTime, counter)

		t := time.NewTimer(backoff)
		select {
		case <-t.C:
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		}

		counter++
	}

	return nil
}

// https://aws.amazon.com/cn/blogs/architecture/exponential-backoff-and-jitter/
func jitterBackoff(min, max time.Duration, attempt int) time.Duration {
	res := deCorrelatedJitter(min, max, attempt)
	if res < min {
		return min
	}

	return res
}

// sleep := tmp / 2 + rand_between(0,tmp/2)
func deCorrelatedJitter(min, max time.Duration, attempt int) time.Duration {

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
func deCorrelatedJitterMoreRound(min, max time.Duration, attempt int) time.Duration {

	base, capV := float64(min), float64(max)
	// tmp := min(cap,base * 2 ** attempt)
	tmp := math.Min(capV, base*math.Exp2(float64(attempt)))

	d := tmp / 2
	jitter := random.Int63n(int64(d))
	sleepBase := int64(d) + jitter
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
