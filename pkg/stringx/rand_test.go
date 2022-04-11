package stringx

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRand(t *testing.T) {
	var (
		randMap = make(map[string]struct{})
		wg      sync.WaitGroup
		mutex   sync.Mutex
	)
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			mutex.Lock()
			tmp := Rand(4)
			//fmt.Println(tmp)
			randMap[tmp] = struct{}{}
			mutex.Unlock()
		}()

	}

	wg.Wait()
	assert.Equal(t, 100, len(randMap))
}
