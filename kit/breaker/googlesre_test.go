package breaker

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testBucketSize = 10
	testBucketTime = time.Millisecond * 10
)

func getTestGoogleSre() *googleSre {
	return newGoogleSre(withGoogleSreBucket(testBucketSize), withGoogleSreWindow(10*testBucketTime))
}

func TestGoogleSre_With(t *testing.T) {
	t.Run("withGoogleSreWindow", func(t *testing.T) {
		sre := newGoogleSre(withGoogleSreWindow(time.Second * 999))
		assert.Equal(t, time.Second*999, sre.option.windowTime)
	})

	t.Run("withGoogleSreBucket", func(t *testing.T) {
		sre := newGoogleSre(withGoogleSreBucket(2233))
		assert.Equal(t, 2233, sre.option.bucketCount)
	})
}

func TestGoogleSre_OFF(t *testing.T) {
	sre := getTestGoogleSre()
	for i := 0; i < 100; i++ {
		err := sre.Allow()
		assert.Nil(t, err)
		sre.MarkSuccess()
	}
	accepts, total := sre.stat()
	assert.True(t, accepts == 100 && total == 100)
	time.Sleep(time.Millisecond * 20)
	for i := 0; i < 60; i++ {
		err := sre.Allow()
		assert.Nil(t, err)
		sre.MarkSuccess()
	}
	accepts, total = sre.stat()
	assert.True(t, accepts == 160 && total == 160)
}

func TestGoogleSre_On(t *testing.T) {
	sre := getTestGoogleSre()
	for i := 0; i < 10; i++ {
		err := sre.Allow()
		assert.Nil(t, err)
		sre.MarkSuccess()
	}
	accepts, total := sre.stat()
	assert.True(t, accepts == 10 && total == 10)

	for i := 0; i < 100000; i++ {
		err := sre.Allow()
		if err == nil {
			sre.MarkFail()
		}
	}

	time.Sleep(testBucketTime * 2)
	drops := 0
	for i := 0; i < 100; i++ {
		if err := sre.Allow(); err != nil {
			drops++
		}
	}
	accepts, total = sre.stat()
	fmt.Println(accepts, total, drops)
	assert.True(t, drops >= 80)

}

func TestGoogleSre_DoReq_Breaker_On(t *testing.T) {
	sre := getTestGoogleSre()
	for i := 0; i < 10; i++ {
		err := sre.Allow()
		assert.Nil(t, err)
		sre.MarkSuccess()
	}
	accepts, total := sre.stat()
	assert.True(t, accepts == 10 && total == 10)

	for i := 0; i < 100000; i++ {
		err := sre.Allow()
		if err == nil {
			sre.MarkFail()
		}
	}

	time.Sleep(testBucketTime)
	err := sre.doReq(func() error {
		return ErrGoogleSreBreakOn
	}, func(err error) error {
		return err
	}, func(err error) bool {
		return err == nil
	})

	assert.ErrorIs(t, err, ErrGoogleSreBreakOn)
}

func TestGoogleSre_DoReq_Breaker_Off(t *testing.T) {
	t.Run("accepted err", func(t *testing.T) {
		sre := getTestGoogleSre()
		acceptErr := errors.New("accepted err")
		err := sre.doReq(func() error {
			return acceptErr
		}, nil, func(err error) bool {
			return errors.Is(err, acceptErr)
		})
		accept, total := sre.stat()
		assert.EqualValues(t, 1, accept)
		assert.EqualValues(t, 1, total)
		assert.ErrorIs(t, err, acceptErr)
	})

	t.Run("unaccepted err", func(t *testing.T) {
		sre := getTestGoogleSre()
		unacceptedErr := errors.New("unaccepted err")
		err := sre.doReq(func() error {
			return unacceptedErr
		}, nil, func(err error) bool {
			return !errors.Is(err, unacceptedErr)
		})
		accept, total := sre.stat()
		assert.EqualValues(t, 0, accept)
		assert.EqualValues(t, 1, total)
		assert.ErrorIs(t, err, unacceptedErr)
	})

	t.Run("panic", func(t *testing.T) {
		sre := getTestGoogleSre()
		assert.Panics(t, func() {
			_ = sre.doReq(func() error {
				panic("got panic")
			}, nil, func(err error) bool {
				return err == nil
			})
		})
	})
}

func TestGoogleBreakerHistory(t *testing.T) {
	var b *googleSre
	var accepts, total int64

	sleep := testBucketTime
	t.Run("accepts == total", func(t *testing.T) {
		b = getTestGoogleSre()
		markSuccessWithDuration(b, 10, sleep/2)
		accepts, total = b.stat()
		assert.Equal(t, int64(10), accepts)
		assert.Equal(t, int64(10), total)
	})

	t.Run("fail == total", func(t *testing.T) {
		b = getTestGoogleSre()
		markFailedWithDuration(b, 10, sleep/2)
		accepts, total = b.stat()
		assert.Equal(t, int64(0), accepts)
		assert.Equal(t, int64(10), total)
	})

	t.Run("accepts = 1/2 * total, fail = 1/2 * total", func(t *testing.T) {
		b = getTestGoogleSre()
		markFailedWithDuration(b, 5, sleep/2)
		markSuccessWithDuration(b, 5, sleep/2)
		accepts, total = b.stat()
		assert.Equal(t, int64(5), accepts)
		assert.Equal(t, int64(10), total)
	})

	t.Run("auto reset rolling counter", func(t *testing.T) {
		b = getTestGoogleSre()
		time.Sleep(testBucketTime * testBucketSize)
		accepts, total = b.stat()
		assert.Equal(t, int64(0), accepts)
		assert.Equal(t, int64(0), total)
	})
}

func markSuccessWithDuration(b *googleSre, count int, sleep time.Duration) {
	for i := 0; i < count; i++ {
		b.MarkSuccess()
		time.Sleep(sleep)
	}
}

func markFailedWithDuration(b *googleSre, count int, sleep time.Duration) {
	for i := 0; i < count; i++ {
		b.MarkFail()
		time.Sleep(sleep)
	}
}

// BenchmarkGoogleBreakerAllow-8   	 2348444	       496.8 ns/op
func BenchmarkGoogleBreakerAllow(b *testing.B) {
	breaker := getTestGoogleSre()
	b.ResetTimer()
	for i := 0; i <= b.N; i++ {
		_ = breaker.Allow()
		if i%2 == 0 {
			breaker.MarkSuccess()
		} else {
			breaker.MarkFail()
		}
	}
}
