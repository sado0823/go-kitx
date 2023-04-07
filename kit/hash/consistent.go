package hash

import (
	"sync"

	"github.com/spaolacci/murmur3"
)

const (
	maxWeight  = 100      // 最大权重参数
	minVirtual            // 最小虚拟节点数量
	prime      = 16777619 // 质数, 减少hash冲突
)

type (
	Consistent interface {
		Add(node interface{}, opts ...ConsistentAddWith)
		Get(v interface{}) (node interface{}, has bool)
		Remove(node interface{})
	}

	consistent struct {
		opt     *option
		vtrNum  int64                    // 虚拟节点数量
		vtrKeys []uint64                 // 虚拟节点值
		vtrRing map[uint64][]interface{} // 虚拟节点环, 每一个值内存储对应真实节点数据
		nodes   map[string]struct{}      // 真实节点
		lock    sync.RWMutex
	}
)

type (
	option struct {
		vtr  int64
		hash func([]byte) uint64
	}

	ConsistentWith func(opt *option)
)

func ConsistentWithVtr(num int64) ConsistentWith {
	return func(opt *option) {
		opt.vtr = num
	}
}

func ConsistentWithHash(hash func([]byte) uint64) ConsistentWith {
	return func(opt *option) {
		opt.hash = hash
	}
}

type (
	optionAdd struct {
		vtr    int64
		weight int64
	}
	ConsistentAddWith func(opt *optionAdd)
)

func ConsistentAddWithVtr(num int64) ConsistentAddWith {
	return func(opt *optionAdd) {
		opt.vtr = num
	}
}

func ConsistentAddWithWeight(weight int64) ConsistentAddWith {
	return func(opt *optionAdd) {
		opt.weight = weight
	}
}

func NewConsistent(withs ...ConsistentWith) Consistent {
	dft := &option{
		vtr:  minVirtual,
		hash: murmur3.Sum64,
	}

	for i := range withs {
		withs[i](dft)
	}

	if dft.vtr < minVirtual {
		dft.vtr = minVirtual
	}

	if dft.hash == nil {
		dft.hash = murmur3.Sum64
	}

	return &consistent{opt: dft}
}

func (c *consistent) Add(node interface{}, opts ...ConsistentAddWith) {

}

func (c *consistent) Get(v interface{}) (node interface{}, has bool) {

}

func (c *consistent) Remove(node interface{}) {

}
