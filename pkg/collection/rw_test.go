package collection

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_Add(t *testing.T) {
	const (
		size     = 5
		interval = time.Millisecond * 1000
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

	assert.Equal(t, []float64{0, 0, 0, 0, 0}, traverse())

	rw.Add(1)
	assert.Equal(t, []float64{0, 0, 0, 1, 0}, traverse())

	wait(time.Millisecond * 50)
	rw.Add(2)
	rw.Add(3)
	assert.Equal(t, []float64{0, 0, 0, 6, 0}, traverse())

	wait(time.Millisecond * 200)
	rw.Add(4)
	rw.Add(5)
	rw.Add(6)
	assert.Equal(t, []float64{0, 0, 0, 6, 15}, traverse())

	wait(time.Millisecond * 200)
	rw.Add(7)
	assert.Equal(t, []float64{7, 0, 0, 6, 15}, traverse())

	wait(time.Millisecond * 200)
	rw.Add(22)
	assert.Equal(t, []float64{7, 22, 0, 6, 15}, traverse())

	wait(time.Millisecond * 200)
	rw.Add(33)
	assert.Equal(t, []float64{7, 22, 33, 6, 15}, traverse())

	wait(time.Millisecond * 200)
	rw.Add(11)
	assert.Equal(t, []float64{7, 22, 33, 11, 15}, traverse())

	wait(time.Millisecond * 400)
	rw.Add(66)
	assert.Equal(t, []float64{7, 66, 33, 11, 0}, traverse())

	wait(time.Millisecond * 600)
	rw.Add(998)
	assert.Equal(t, []float64{0, 66, 0, 11, 998}, traverse())
}

//func Test_Add(t *testing.T) {
//	const (
//		size     = 3
//		interval = time.Millisecond * 50
//	)
//
//	rw := New(size, interval)
//	traverse := func() []float64 {
//		var sum []float64
//		rw.Reduce(func(bucket *Bucket) {
//			sum = append(sum, bucket.Sum)
//		})
//		return sum
//	}
//	wait := func() {
//		time.Sleep(interval)
//	}
//
//	assert.Equal(t, []float64{0, 0, 0}, traverse())
//
//	rw.Add(1)
//	assert.Equal(t, []float64{0, 0, 1}, traverse())
//
//	wait()
//	rw.Add(2)
//	rw.Add(3)
//	assert.Equal(t, []float64{0, 1, 5}, traverse())
//
//	wait()
//	rw.Add(4)
//	rw.Add(5)
//	rw.Add(6)
//	assert.Equal(t, []float64{1, 5, 15}, traverse())
//
//	wait()
//	rw.Add(7)
//	assert.Equal(t, []float64{5, 15, 7}, traverse())
//
//}
