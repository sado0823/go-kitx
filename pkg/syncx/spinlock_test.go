package syncx

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpinLock_TryLock(t *testing.T) {
	t.Run("try lock directly", func(t *testing.T) {
		var lock SpinLock
		assert.True(t, lock.TryLock())
		assert.False(t, lock.TryLock())
		lock.Unlock()
		assert.True(t, lock.TryLock())
	})

	t.Run("lock first", func(t *testing.T) {
		var lock SpinLock
		lock.Lock()
		assert.False(t, lock.TryLock())
		lock.Unlock()
		assert.True(t, lock.TryLock())
	})

}

func TestSpinLock_Race(t *testing.T) {
	var lock SpinLock
	var count int32
	var wait sync.WaitGroup
	wait.Add(2)
	sig := make(chan struct{})

	go func() {
		lock.TryLock()
		sig <- struct{}{}
		atomic.AddInt32(&count, 1)
		runtime.Gosched()
		lock.Unlock()
		wait.Done()
	}()

	go func() {
		<-sig
		lock.Lock()
		atomic.AddInt32(&count, 1)
		lock.Unlock()
		wait.Done()
	}()

	wait.Wait()
	assert.Equal(t, int32(2), atomic.LoadInt32(&count))
}

func BenchmarkSpinLock(b *testing.B) {
	b.Run("spin lock", func(b *testing.B) {
		var spin = &SpinLock{}
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				spin.Lock()
			} else {
				spin.Unlock()
			}
		}
	})
	b.Run("mutex", func(b *testing.B) {
		b.ResetTimer()
		var mutex = sync.Mutex{}
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				mutex.Lock()
			} else {
				mutex.Unlock()
			}
		}
	})
}
