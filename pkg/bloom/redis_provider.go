package bloom

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/sado0823/go-kitx/kit/store/redis"
	"github.com/spaolacci/murmur3"
)

const (
	// for detail, see http://pages.cs.wisc.edu/~cao/papers/summary-cache/node8.html
	maps      = 14
	setScript = `
for _, offset in ipairs(ARGV) do
	redis.call("setbit", KEYS[1], offset, 1)
end
`
	checkScript = `
for _, offset in ipairs(ARGV) do
	if tonumber(redis.call("getbit", KEYS[1], offset)) == 0 then
		return false
	end
end
return true
`
)

var ErrTooLargeOffset = errors.New("too large offset")

type rdsProvider struct {
	store *redis.Redis
	key   string
	bits  uint
}

func NewRedisProvider(addr string, key string, bits uint) Provider {
	return &rdsProvider{store: redis.New(addr), key: key, bits: bits}
}

// Add implement Provider interface
func (r *rdsProvider) Add(ctx context.Context, data []byte) error {
	location := r.getBitLocation(data)
	return r.set(ctx, location)
}

// Exists implement Provider interface
func (r *rdsProvider) Exists(ctx context.Context, data []byte) (bool, error) {
	location := r.getBitLocation(data)
	return r.check(ctx, location)
}

// getBitLocation return data hash to bit location
func (r *rdsProvider) getBitLocation(data []byte) []uint {
	l := make([]uint, maps)
	for i := 0; i < maps; i++ {
		hashV := r.hash(append(data, byte(i)))
		l[i] = uint(hashV % uint64(maps))
	}
	return l
}

// set those offsets into bloom filter
func (r *rdsProvider) set(ctx context.Context, offsets []uint) error {
	args, err := r.buildOffsetArgs(offsets)
	if err != nil {
		return err
	}

	_, err = r.store.Eval(ctx, setScript, []string{r.key}, args)
	if errors.Is(err, redis.Nil) {
		return nil
	}

	return err
}

// check if those offsets are in bloom filter
func (r *rdsProvider) check(ctx context.Context, offsets []uint) (bool, error) {
	args, err := r.buildOffsetArgs(offsets)
	if err != nil {
		return false, err
	}

	eval, err := r.store.Eval(ctx, checkScript, []string{r.key}, args)
	if errors.Is(err, redis.Nil) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return fmt.Sprintf("%v", eval) == "1", nil
}

// buildOffsetArgs set []uint offset to []string that can use in redis
// and check if offset is larger than r.bits
func (r *rdsProvider) buildOffsetArgs(offsets []uint) ([]string, error) {
	var args []string

	for _, offset := range offsets {
		if offset >= r.bits {
			return nil, ErrTooLargeOffset
		}

		args = append(args, strconv.FormatUint(uint64(offset), 10))

	}

	return args, nil
}

// hash returns the hash value of data.
func (r *rdsProvider) hash(data []byte) uint64 {
	return murmur3.Sum64(data)
}
