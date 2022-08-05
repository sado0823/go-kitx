package v1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	assert.NotNil(t, New(10, time.Second))
	assert.Panics(t, func() {
		New(0, time.Second)
	})
}

func Test_Add(t *testing.T) {
	const (
		size     = 3
		interval = time.Millisecond * 100
	)

	rw := New(size, interval)
	traverse := func() []float64 {
		var sum []float64
		rw.Reduce(func(bucket *Bucket) {
			sum = append(sum, bucket.Sum)
		})
		return sum
	}
	wait := func(interval time.Duration) {
		time.Sleep(interval)
	}

	assert.Equal(t, []float64{0, 0, 0}, traverse())

	rw.Add(1)
	assert.Equal(t, []float64{1, 0, 0}, traverse())

	wait(time.Millisecond * 50) // 50ms
	rw.Add(2)
	rw.Add(3)
	assert.Equal(t, []float64{6, 0, 0}, traverse())

	wait(time.Millisecond * 100) // 150ms
	rw.Add(4)
	rw.Add(5)
	rw.Add(6)
	assert.Equal(t, []float64{6, 15, 0}, traverse())

	wait(time.Millisecond * 100) // 250ms
	rw.Add(7)
	rw.Add(8)
	assert.Equal(t, []float64{6, 15, 15}, traverse())

	wait(time.Millisecond * 100) // 350ms
	rw.Add(9)
	assert.Equal(t, []float64{9, 15, 15}, traverse())

	wait(time.Millisecond * 10) // 360ms
	rw.Add(9)
	assert.Equal(t, []float64{18, 15, 15}, traverse())

	wait(time.Millisecond * 300) // 660ms
	assert.Equal(t, []float64{0, 0, 0}, traverse())

	wait(time.Millisecond * 100) // 760ms
	rw.Add(1)
	assert.Equal(t, []float64{0, 1, 0}, traverse())

	wait(interval) // 860ms
	assert.Equal(t, []float64{0, 1, 0}, traverse())

	wait(interval) // 960ms
	assert.Equal(t, []float64{0, 1, 0}, traverse())
}

func Test_Reduce(t *testing.T) {
	const (
		size     = 4
		interval = time.Millisecond * 50
	)
	wait := func() {
		time.Sleep(interval)
	}

	rw := New(size, interval)
	// bucket0: 0 = 0 (sleep)
	// bucket1: 0+1 = 1 (sleep)
	// bucket2: 0+1+2 = 3 (sleep)
	// bucket3: 0+1+2+3 = 6
	for x := 0; x < size; x++ {
		for i := 0; i <= x; i++ {
			rw.Add(float64(i))
		}
		// 0,1,2
		if x < size-1 {
			wait()
		}
	}
	res := float64(0)
	rw.Reduce(func(bucket *Bucket) {
		res += bucket.Sum
	})
	assert.EqualValues(t, 10, res)
}

// Benchmark_RW-8   	 1136754	      1039 ns/op
func Benchmark_RW(b *testing.B) {
	const (
		size     = 4
		interval = time.Millisecond * 50
	)
	rw := New(size, interval)
	b.ResetTimer()
	for i := 0; i <= b.N; i++ {
		if i%2 == 0 {
			rw.Add(float64(i))
		} else if i%100 == 0 {
			rw.Reduce(func(bucket *Bucket) {
				_ = bucket.Sum + float64(i)
			})
		} else if i%500 == 0 {
			time.Sleep(time.Millisecond * 20)
		}
	}
}
