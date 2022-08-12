package ratelimit

import (
	"context"
	"errors"
	"github.com/sado0823/go-kitx/kit/store/redis"
	"strconv"
	"time"
)

const periodScript = `local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local current = redis.call("INCRBY", KEYS[1], 1)
if current == 1 then
	redis.call("EXPIRE", KEYS[1], window)
end
if current < limit then
	return 1
elseif current == limit then
	return 2
else
	return 0
end`

type PeriodCode int

const (
	PeriodUnknown PeriodCode = iota
	PeriodAllowed
	PeriodHitQuota
	PeriodOverQuota

	periodAllowed   = 1
	periodHitQuota  = 2
	periodOverQuota = 0
)

var ErrUnknownPeriodCode = errors.New("unknown period code")

type (
	PeriodOption func(p *Period)

	Period struct {
		period    int
		quota     int
		store     *redis.Redis
		keyPrefix string
		align     bool
	}
)

func WithAlign() PeriodOption {
	return func(p *Period) {
		p.align = true
	}
}

func WithPrefix(prefix string) PeriodOption {
	return func(p *Period) {
		p.keyPrefix = prefix
	}
}

func NewPeriod(seconds, quota int, store *redis.Redis, options ...PeriodOption) *Period {
	limiter := &Period{
		period: seconds,
		quota:  quota,
		store:  store,
	}
	for _, periodOption := range options {
		periodOption(limiter)
	}
	return limiter
}

func (p *Period) TakeCtx(ctx context.Context, key string) (PeriodCode, error) {
	resp, err := p.store.EvalCtx(ctx, periodScript, []string{p.keyPrefix + key}, []string{
		strconv.Itoa(p.quota),
		strconv.Itoa(p.calcExpireSeconds()),
	})
	if err != nil {
		return PeriodUnknown, err
	}

	code, ok := resp.(int64)
	if !ok {
		return PeriodUnknown, ErrUnknownPeriodCode
	}

	switch code {
	case periodOverQuota:
		return PeriodOverQuota, nil
	case periodAllowed:
		return PeriodAllowed, nil
	case periodHitQuota:
		return PeriodHitQuota, nil
	default:
		return PeriodUnknown, ErrUnknownPeriodCode
	}
}

func (p *Period) calcExpireSeconds() int {
	if p.align {
		now := time.Now()
		_, offset := now.Zone()
		unix := now.Unix() + int64(offset)
		return p.period - int(unix%int64(p.period))
	}
	return p.period
}
