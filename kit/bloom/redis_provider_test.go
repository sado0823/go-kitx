package bloom

import (
	"context"
	"testing"
	"time"

	"github.com/sado0823/go-kitx/kit/store/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

// createRedis returns an in process redis.Redis.
func createRedis() (addr string, clean func(), err error) {
	mr, err := miniredis.Run()
	if err != nil {
		return "", nil, err
	}

	return mr.Addr(), func() {
		ch := make(chan struct{})
		go func() {
			mr.Close()
			close(ch)
		}()
		select {
		case <-ch:
		case <-time.After(time.Second):
		}
	}, nil
}

func TestRedisBitSet_New_Set_Test(t *testing.T) {
	addr, clean, err := createRedis()
	assert.Nil(t, err)
	defer clean()
	ctx := context.Background()

	bitSet := &rdsProvider{store: redis.New(addr), key: "test_key", bits: 1024}
	isSetBefore, err := bitSet.check(ctx, []uint{0})
	if err != nil {
		t.Fatal(err)
	}
	if isSetBefore {
		t.Fatal("Bit should not be set")
	}
	err = bitSet.set(ctx, []uint{512})
	if err != nil {
		t.Fatal(err)
	}
	isSetAfter, err := bitSet.check(ctx, []uint{512})
	if err != nil {
		t.Fatal(err)
	}
	if !isSetAfter {
		t.Fatal("Bit should be set")
	}

}

func TestRedisBitSet_Add(t *testing.T) {
	addr, clean, err := createRedis()
	assert.Nil(t, err)
	defer clean()

	ctx := context.Background()

	filter := &rdsProvider{store: redis.New(addr), key: "test_key", bits: 1024}
	assert.Nil(t, filter.Add(ctx, []byte("hello")))
	assert.Nil(t, filter.Add(ctx, []byte("world")))
	ok, err := filter.Exists(ctx, []byte("hello"))
	assert.Nil(t, err)
	assert.True(t, ok)
}
