package v2

import (
	"testing"
	"time"

	"github.com/sado0823/go-kitx/pkg/stringx"

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
		interval = time.Millisecond * 50
	)

	rw := New(size, interval)
	traverse := func() []float64 {
		var sum []float64
		rw.Reduce(func(bucket *Bucket) {
			sum = append(sum, bucket.Sum)
		})
		return sum
	}
	wait := func() {
		time.Sleep(interval)
	}

	assert.Equal(t, []float64{0, 0, 0}, traverse())

	rw.Add(1)
	assert.Equal(t, []float64{0, 0, 1}, traverse())

	wait()
	rw.Add(2)
	rw.Add(3)
	assert.Equal(t, []float64{0, 1, 5}, traverse())

	wait()
	rw.Add(4)
	rw.Add(5)
	rw.Add(6)
	assert.Equal(t, []float64{1, 5, 15}, traverse())

	wait()
	rw.Add(7)
	assert.Equal(t, []float64{5, 15, 7}, traverse())

	time.Sleep(interval * 2)
	rw.Add(123)
	assert.Equal(t, []float64{7, 0, 123}, traverse())

	time.Sleep(interval * 3)
	assert.Nil(t, traverse())

}

func Test_Reduce(t *testing.T) {
	const (
		size     = 4
		interval = time.Millisecond * 50
	)
	wait := func() {
		time.Sleep(interval)
	}
	cases := []struct {
		win    *RollingWindow
		expect float64
	}{
		{
			win:    New(size, interval),
			expect: 10,
		},
		{
			win:    New(size, interval, WithIgnoreCurrent()),
			expect: 4,
		},
	}

	// 0
	// 0 1
	// 0 1 2
	// 0 1 2 3
	for _, caseV := range cases {
		t.Run(stringx.Rand(6), func(t *testing.T) {
			r := caseV.win
			for x := 0; x < size; x++ {
				for i := 0; i <= x; i++ {
					r.Add(float64(i))
				}
				// 0,1,2
				if x < size-1 {
					wait()
				}
			}
			var result float64
			r.Reduce(func(b *Bucket) {
				result += b.Sum
			})
			assert.Equal(t, caseV.expect, result)
		})
	}
}

// Benchmark_RollingWindow-8   	17171872	        70.30 ns/op
func Benchmark_RollingWindow(b *testing.B) {
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
