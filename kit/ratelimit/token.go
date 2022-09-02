package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sado0823/go-kitx/kit/store/redis"

	xrate "golang.org/x/time/rate"
)

const (
	// KEYS[1] token_key
	// KEYS[2] timestamp_key
	tokenScript = `local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])
local fill_time = capacity/rate
local ttl = math.floor(fill_time*2)
local last_tokens = tonumber(redis.call("get",KEYS[1]))
if last_tokens == nil then
	last_tokens = capacity
end

local last_refreshed = tonumber(redis.call("get",KEYS[2]))
if last_refreshed == nil then
	last_refreshed = 0
end

local delta = math.max(0,now-last_refreshed)
local filled_tokens = math.min(capacity,(delta*rate)+last_tokens)
local allowed = filled_tokens >= requested
local new_tokens = filled_tokens
if allowed then
	new_tokens = filled_tokens - requested
end

redis.call("setex", KEYS[1], ttl, new_tokens)
redis.call("setex", KEYS[2], ttl, now)

return allowed`

	tokenFormat          = "{%s}.token"
	tokenTimestampFormat = "{%s}.timestamp"
	tokenPingInterval    = time.Microsecond * 100
)

type Token struct {
	rate         int
	capacity     int
	store        *redis.Redis
	tokenKey     string
	timestampKey string
	lock         sync.Mutex
	storeAlive   uint32
	degrade      *xrate.Limiter
	monitorOn    bool
}

func NewToken(rate, capacity int, store *redis.Redis, key string) *Token {
	if rate <= 0 || capacity <= 0 || key == "" || store == nil {
		panic("invalid token param")
	}

	tokenKey := fmt.Sprintf(tokenFormat, key)
	timestampKey := fmt.Sprintf(tokenTimestampFormat, key)

	return &Token{
		rate:         rate,
		capacity:     capacity,
		store:        store,
		tokenKey:     tokenKey,
		timestampKey: timestampKey,
		storeAlive:   1,
		degrade:      xrate.NewLimiter(xrate.Every(time.Second/time.Duration(rate)), capacity),
	}
}

func (t *Token) Allow(ctx context.Context) bool {
	return t.AllowN(ctx, time.Now(), 1)
}

func (t *Token) AllowN(ctx context.Context, now time.Time, n int) bool {
	return t.reserveN(ctx, now, n)
}

func (t *Token) reserveN(ctx context.Context, now time.Time, n int) bool {
	select {
	case <-ctx.Done():
		logger.Printf("fail to use rate limiter: %s", ctx.Err())
		return false
	default:
	}

	if atomic.LoadUint32(&t.storeAlive) == 0 {
		return t.degrade.AllowN(now, n)
	}

	eval, err := t.store.Eval(ctx, tokenScript,
		[]string{t.tokenKey, t.timestampKey},
		[]string{
			strconv.Itoa(t.rate),
			strconv.Itoa(t.capacity),
			strconv.FormatInt(now.Unix(), 10),
			strconv.Itoa(n),
		},
	)

	// redis allowed == false
	// Lua boolean false -> r Nil bulk reply
	if err == redis.Nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		logger.Printf("fail to use rate limiter: %s", err)
		return false
	}

	if err != nil {
		logger.Printf("token limit eval err:%+v,resp:%v, use in-process limit instead", err, eval)
		t.startMonitor()
		return t.degrade.AllowN(now, n)
	}

	code, ok := eval.(int64)
	if !ok {
		logger.Printf("token limit eval err:%+v,resp:%v, use in-process limit instead", err, eval)
		t.startMonitor()
		return t.degrade.AllowN(now, n)
	}

	// redis allowed == true
	// Lua boolean true -> r integer reply with value of 1
	return code == 1
}

func (t *Token) startMonitor() {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.monitorOn {
		return
	}

	t.monitorOn = true
	atomic.StoreUint32(&t.storeAlive, 0)

	go t.heartbeat()
}

func (t *Token) heartbeat() {
	ticker := time.NewTicker(tokenPingInterval)
	defer func() {
		ticker.Stop()
		t.lock.Lock()
		t.monitorOn = false
		t.lock.Unlock()
	}()

	for range ticker.C {
		if err := t.store.Ping(context.Background()); err == nil {
			atomic.StoreUint32(&t.storeAlive, 1)
			return
		}
	}
}
