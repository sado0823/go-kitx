package rollingwindow

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

}
