package hash

import (
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewConsistent(t *testing.T) {
	assert.NotPanics(t, func() {
		NewConsistent()
	})

	assert.NotPanics(t, func() {
		NewConsistent(ConsistentWithHash(nil))
	})

	assert.NotPanics(t, func() {
		NewConsistent(ConsistentWithVtr(-1))
	})
}

func Test_Consistent_Get(t *testing.T) {
	ch := NewConsistent()
	for i := 0; i < 20; i++ {
		ch.Add("prefix" + strconv.Itoa(i))
	}

	keys := make(map[int]string, 1000)
	for i := 0; i < 1000; i++ {
		key, ok := ch.Get(1000 + i)
		assert.True(t, ok)
		assert.NotNil(t, key)
		keys[i] = key.(string)
	}

}

func Test_Consistent_Cascade(t *testing.T) {
	hash := NewConsistent(ConsistentWithVtr(10))

	cascade := hash.Cascade("foo")
	t.Logf("cascade: %v", cascade)
	assert.True(t, len(cascade) == 100)
}

func TestConsistentHash(t *testing.T) {
	ch := NewConsistent()
	val, ok := ch.Get("any")
	assert.False(t, ok)
	assert.Nil(t, val)

	for i := 0; i < 20; i++ {
		ch.Add("localhost:"+strconv.Itoa(i), ConsistentAddWithVtr(minVirtual<<1))
	}

	keys := make(map[string]int)
	for i := 0; i < 1000; i++ {
		key, ok := ch.Get(1000 + i)
		assert.True(t, ok)
		keys[key.(string)]++
	}

	mi := make(map[interface{}]int, len(keys))
	for k, v := range keys {
		mi[k] = v
	}
	entropy := calcEntropy(mi)
	assert.True(t, entropy > .95)
}

const epsilon = 1e-6

func calcEntropy(m map[interface{}]int) float64 {
	if len(m) == 0 || len(m) == 1 {
		return 1
	}

	var entropy float64
	var total int
	for _, v := range m {
		total += v
	}

	for _, v := range m {
		proba := float64(v) / float64(total)
		if proba < epsilon {
			proba = epsilon
		}
		entropy -= proba * math.Log2(proba)
	}

	return entropy / math.Log2(float64(len(m)))
}
