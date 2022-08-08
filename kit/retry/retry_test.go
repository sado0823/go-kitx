package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Loop(t *testing.T) {
	t.Run("ctx timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
		defer cancel()
		err := Loop(ctx, "", func(ctx context.Context) error {
			return errors.New("ctx timeout")
		})
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

func Test_Func(t *testing.T) {
	t.Run("ctx timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
		defer cancel()
		err := Func(ctx, "", func(ctx context.Context) error {
			return errors.New("ctx timeout")
		})
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("ctx cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		err := Func(ctx, "", func(ctx context.Context) error {
			cancel()
			return errors.New("ctx cancel")
		})
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("with max attempt", func(t *testing.T) {
		var (
			errSet     = errors.New("with max attempt")
			maxAttempt = 5
			tries      = 0
		)
		err := Func(context.Background(), "", func(ctx context.Context) error {
			tries++
			return errSet
		}, WithLimit(maxAttempt))
		fmt.Println(err)
		assert.ErrorIs(t, err, errSet)
		assert.Equal(t, maxAttempt, tries-1)
	})

	t.Run("with retry success", func(t *testing.T) {
		var (
			errSet      = errors.New("with retry success")
			maxAttempt  = 5
			tries       = 0
			withSuccess = 3
		)
		err := Func(context.Background(), "", func(ctx context.Context) error {
			tries++
			if tries == withSuccess {
				return nil
			}
			return errSet
		}, WithLimit(maxAttempt), WithMax(time.Millisecond*500))
		fmt.Println(err)
		assert.Nil(t, err)
		assert.Equal(t, withSuccess, tries)
	})
}

func Test_jitter(t *testing.T) {

	var (
		defaultMaxRetries  = 3
		defaultWaitTime    = time.Duration(100) * time.Millisecond
		defaultMaxWaitTime = time.Duration(2000) * time.Millisecond
	)

	t.Run("v1", func(t *testing.T) {
		var total time.Duration = 0
		for attempt := 0; attempt <= defaultMaxRetries; attempt++ {
			v := deCorrelatedJitter(defaultWaitTime, defaultMaxWaitTime, attempt)
			total += v
			fmt.Printf("v1, attempt:%d, wait:%v \n", attempt, v)
			assert.Equal(t, true, v <= defaultMaxWaitTime)
		}
		fmt.Println("v1 total:", total)
	})

	t.Run("v2", func(t *testing.T) {
		var total time.Duration = 0
		for attempt := 0; attempt <= defaultMaxRetries; attempt++ {
			v := deCorrelatedJitterMoreRound(defaultWaitTime, defaultMaxWaitTime, attempt)
			total += v
			fmt.Printf("v2, attempt:%d, wait:%v \n", attempt, v)
			assert.Equal(t, true, v <= defaultMaxWaitTime)
		}
		fmt.Println("v2 total:", total)
	})

}
