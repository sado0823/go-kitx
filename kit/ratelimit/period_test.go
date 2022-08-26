package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/sado0823/go-kitx/kit/store/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func Test_Period_Run(t *testing.T) {
	var (
		seconds int = 1
		quota   int = 5
		total       = 100
		ctx         = context.Background()
		key         = "run"
	)

	mini, err := miniredis.Run()
	assert.Nil(t, err)
	defer mini.Close()

	t.Run("not align", func(t *testing.T) {
		period := NewPeriod(seconds, quota, redis.New(mini.Addr()), WithPeriodPrefix("period"))
		var allowed, hit, over int
		for i := 0; i < total; i++ {
			take, err := period.Take(ctx, key)
			assert.Nil(t, err)
			switch take {
			case PeriodAllowed:
				allowed++
			case PeriodHitQuota:
				hit++
			case PeriodOverQuota:
				over++
			default:
				t.Errorf("unknown period code:%d", take)
			}
		}
		assert.Equal(t, quota-1, allowed)
		assert.Equal(t, 1, hit)
		assert.Equal(t, total-quota, over)
	})

	t.Run("with align", func(t *testing.T) {
		period := NewPeriod(seconds, quota, redis.New(mini.Addr()), WithPeriodPrefix("period-align"), WithPeriodAlign())
		var allowed, hit, over int
		for i := 0; i < total; i++ {
			take, err := period.Take(ctx, key)
			assert.Nil(t, err)
			switch take {
			case PeriodAllowed:
				allowed++
			case PeriodHitQuota:
				hit++
			case PeriodOverQuota:
				over++
			default:
				t.Errorf("unknown period code:%d", take)
			}
		}
		assert.Equal(t, quota-1, allowed)
		assert.Equal(t, 1, hit)
		assert.Equal(t, total-quota, over)
	})

}

func Test_Period_Param(t *testing.T) {
	newPeriod := func(t *testing.T, seconds, quota int) (period *Period, store *miniredis.Miniredis, close func()) {
		mini, err := miniredis.Run()
		assert.Nil(t, err)
		period = NewPeriod(seconds, quota, redis.New(mini.Addr()))
		return period, mini, mini.Close
	}

	t.Run("-1 second", func(t *testing.T) {
		period, _, c := newPeriod(t, -1, 1)
		defer c()

		for i := 0; i < 100; i++ {
			take, err := period.Take(context.Background(), "-1 second")
			assert.Nil(t, err)
			assert.Equal(t, PeriodHitQuota, take)
		}
	})

	t.Run("-1 quota", func(t *testing.T) {
		period, _, c := newPeriod(t, 1, -1)
		defer c()

		take, err := period.Take(context.Background(), "-1 quota")
		assert.Nil(t, err)
		assert.Equal(t, PeriodOverQuota, take)
	})

	t.Run("zero second", func(t *testing.T) {
		period, mini, c := newPeriod(t, 0, 1)
		defer c()

		for i := 0; i < 100; i++ {
			// decrease ttl
			mini.FastForward(0)
			take, err := period.Take(context.Background(), "zero second")
			assert.Nil(t, err)
			assert.Equal(t, PeriodHitQuota, take)
		}
	})

	t.Run("zero quota", func(t *testing.T) {
		period, mini, c := newPeriod(t, 1, 0)
		defer c()
		t.Run("ttl expired", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				// decrease ttl
				mini.FastForward(time.Second)
				take, err := period.Take(context.Background(), "ttl expired")
				assert.Nil(t, err)
				assert.Equal(t, PeriodOverQuota, take)
			}
		})

		t.Run("ttl Not expired", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				take, err := period.Take(context.Background(), "ttl Not expired")
				assert.Nil(t, err)
				assert.Equal(t, PeriodOverQuota, take)
			}
		})

	})
}

func Test_Period_Code(t *testing.T) {
	t.Run("quota", func(t *testing.T) {
		var (
			seconds int = 3
			quota   int = 2
			ctx         = context.Background()
			key         = "quota"
		)

		mini, err := miniredis.Run()
		assert.Nil(t, err)
		defer mini.Close()

		period := NewPeriod(seconds, quota, redis.New(mini.Addr()))

		// first allowed
		take, err := period.Take(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, PeriodAllowed, take)

		// second hit full
		take, err = period.Take(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, PeriodHitQuota, take)

		// third hit overload
		take, err = period.Take(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, PeriodOverQuota, take)

		// still overload
		take, err = period.Take(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, PeriodOverQuota, take)
	})

	t.Run("expire", func(t *testing.T) {
		var (
			seconds int = 1
			quota   int = 1
			ctx         = context.Background()
			key         = "quota"
		)

		mini, err := miniredis.Run()
		assert.Nil(t, err)
		defer mini.Close()

		period := NewPeriod(seconds, quota, redis.New(mini.Addr()))
		take, err := period.Take(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, PeriodHitQuota, take)

		take, err = period.Take(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, PeriodOverQuota, take)

		for i := 0; i < 100; i++ {
			// decrease ttl
			mini.FastForward(time.Second * 1)
			take, err = period.Take(ctx, key)
			assert.Nil(t, err)
			assert.Equal(t, PeriodHitQuota, take)
		}
	})

}

func Test_Period_Redis_Fail(t *testing.T) {
	var (
		seconds int = 3
		quota   int = 10
		ctx         = context.Background()
	)

	mini, err := miniredis.Run()
	assert.Nil(t, err)

	period := NewPeriod(seconds, quota, redis.New(mini.Addr()))
	mini.Close()

	take, err := period.Take(ctx, "redis-fail")
	assert.NotNil(t, err)
	t.Log(err)
	assert.Equal(t, PeriodUnknown, take)

}

func Test_Period_Key(t *testing.T) {
	var (
		seconds int = 3
		quota   int = 10
		ctx         = context.Background()
	)

	t.Run("no prefix", func(t *testing.T) {
		mini, err := miniredis.Run()
		assert.Nil(t, err)
		defer mini.Close()

		period := NewPeriod(seconds, quota, redis.New(mini.Addr()))
		take, err := period.Take(ctx, "no-prefix")
		assert.Nil(t, err)
		assert.Equal(t, PeriodAllowed, take)

		in := mini.Exists("no-prefix")
		assert.True(t, in)
	})

	t.Run("with prefix", func(t *testing.T) {
		mini, err := miniredis.Run()
		assert.Nil(t, err)
		defer mini.Close()

		var (
			prefix = "period"
			key    = "with-prefix"
		)

		period := NewPeriod(seconds, quota, redis.New(mini.Addr()), WithPeriodPrefix(prefix))
		take, err := period.Take(ctx, key)
		assert.Nil(t, err)
		assert.Equal(t, PeriodAllowed, take)

		in := mini.Exists(key)
		assert.False(t, in)

		in = mini.Exists(prefix + key)
		assert.True(t, in)

	})
}
