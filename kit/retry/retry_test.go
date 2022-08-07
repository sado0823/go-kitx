package retry

import (
	"fmt"
	"testing"
	"time"
)

func Test_jitter(t *testing.T) {

	var (
		defaultMaxRetries  = 5
		defaultWaitTime    = time.Duration(100) * time.Millisecond
		defaultMaxWaitTime = time.Duration(2000) * time.Millisecond
	)

	t.Run("v1", func(t *testing.T) {
		var total time.Duration = 0
		for attempt := 0; attempt <= defaultMaxRetries; attempt++ {
			v := randDurationV1(defaultWaitTime, defaultMaxWaitTime, attempt)
			total += v
			fmt.Printf("v1, attempt:%d, wait:%v \n", attempt, v)
		}
		fmt.Println("v1 total:", total)
	})

	t.Run("v2", func(t *testing.T) {
		var total time.Duration = 0
		for attempt := 0; attempt <= defaultMaxRetries; attempt++ {
			v := randDurationV2(defaultWaitTime, defaultMaxWaitTime, attempt)
			total += v
			fmt.Printf("v2, attempt:%d, wait:%v \n", attempt, v)
		}
		fmt.Println("v2 total:", total)
	})

}
