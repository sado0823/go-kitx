package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/sado0823/go-kitx/kit/store/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestToken_WithCtx(t *testing.T) {
	s, err := miniredis.Run()
	assert.Nil(t, err)

	const (
		total = 100
		rate  = 5
		burst = 10
	)
	l := NewToken(rate, burst, redis.New(s.Addr()), "token-test")
	defer s.Close()

	ctx, cancel := context.WithCancel(context.Background())
	ok := l.Allow(ctx)
	assert.True(t, ok)

	cancel()
	for i := 0; i < total; i++ {
		ok := l.Allow(ctx)
		assert.False(t, ok)
		assert.False(t, l.monitorOn)
	}
}

func TestToken_Degrade(t *testing.T) {
	mini, err := miniredis.Run()
	assert.Nil(t, err)

	var (
		rate     = 5
		capacity = 10
		total    = 100
		ctx      = context.Background()
	)

	token := NewToken(rate, capacity, redis.New(mini.Addr()), "token-test")
	mini.Close()

	var allowed int
	for i := 0; i < total; i++ {
		time.Sleep(time.Second / time.Duration(total))
		if i == total>>1 {
			assert.Nil(t, mini.Restart())
		}
		if token.Allow(ctx) {
			allowed++
		}

		// start more than once
		token.startMonitor()
	}

	assert.True(t, allowed >= capacity+rate)
}

func TestToken_Take(t *testing.T) {
	mini, err := miniredis.Run()
	assert.Nil(t, err)
	defer mini.Close()

	var (
		rate     = 5
		capacity = 10
		total    = 100
		ctx      = context.Background()
	)

	token := NewToken(rate, capacity, redis.New(mini.Addr()), "token-test")
	var allowed int
	for i := 0; i < total; i++ {
		time.Sleep(time.Second / time.Duration(total))
		if token.Allow(ctx) {
			allowed++
		}
	}

	assert.True(t, allowed >= capacity+rate)
}

func TestToken_TakeBurst(t *testing.T) {
	mini, err := miniredis.Run()
	assert.Nil(t, err)
	defer mini.Close()

	var (
		rate     = 5
		capacity = 10
		total    = 100
		ctx      = context.Background()
	)

	token := NewToken(rate, capacity, redis.New(mini.Addr()), "token-test")
	var allowed int
	for i := 0; i < total; i++ {
		if token.Allow(ctx) {
			allowed++
		}
	}

	assert.True(t, allowed >= capacity)
}
