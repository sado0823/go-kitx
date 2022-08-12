package ratelimit

import (
	"errors"
	"time"
)

const (
	bbrBuckets         = 50
	bbrWindow          = time.Second * 5
	bbrCpuThreshold    = 900
	bbrMinRt           = float64(time.Second / time.Millisecond)
	bbrBeta            = 0.9
	bbrCoolingDuration = time.Second
)

var ErrBBRServiceOverload = errors.New("service overload with bbr")

type (
	BBROptionFn func(*bbrOption)

	bbrOption struct {
		window       time.Duration
		buckets      int
		cpuThreshold int64
	}

	// BBR Bottleneck Bandwidth and RTT
	BBR struct {
		option *bbrOption
		flying int64
		// todo
		// ...
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

	return nil
}
