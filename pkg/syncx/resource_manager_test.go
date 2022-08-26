package syncx

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testRsc struct {
	times int64
}

func (t *testRsc) Close() error {
	return errors.New("testRsc close")
}

func TestResourceManager_Get(t *testing.T) {
	manager := NewResourceManager()
	defer manager.Close()

	var (
		times int64
		wg    sync.WaitGroup
	)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			get, err := manager.Get("key", func() (io.Closer, error) {
				atomic.AddInt64(&times, 1)
				return &testRsc{times: times}, nil
			})
			assert.Nil(t, err)
			assert.Equal(t, int64(1), get.(*testRsc).times)
		}()
	}
	wg.Wait()
}

func TestResourceManager_Get_Error(t *testing.T) {
	manager := NewResourceManager()
	defer manager.Close()

	var e = errors.New("get rsc err")
	for i := 0; i < 10; i++ {
		_, err := manager.Get("key", func() (io.Closer, error) {
			return nil, e
		})
		assert.ErrorIs(t, err, e)
	}
}

func TestResourceManager_Close(t *testing.T) {
	manager := NewResourceManager()
	defer manager.Close()

	var e = errors.New("get rsc err")
	for i := 0; i < 10; i++ {
		_, err := manager.Get("key", func() (io.Closer, error) {
			return nil, e
		})
		assert.ErrorIs(t, err, e)
	}

	assert.NoError(t, manager.Close())
	assert.Equal(t, 0, len(manager.resources))
}

func TestResourceManager_UseAfterClose(t *testing.T) {
	manager := NewResourceManager()
	defer manager.Close()

	var e = errors.New("get rsc err")
	_, err := manager.Get("key", func() (io.Closer, error) {
		return nil, e
	})
	assert.ErrorIs(t, err, e)

	assert.NoError(t, manager.Close())

	_, err = manager.Get("key", func() (io.Closer, error) {
		return nil, e
	})
	assert.ErrorIs(t, err, e)

	assert.Panics(t, func() {
		_, err = manager.Get("key", func() (io.Closer, error) {
			return &testRsc{times: 666}, nil
		})
	})

}
