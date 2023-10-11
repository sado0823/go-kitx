package hash

import (
	"fmt"
	"sort"
	"strconv"
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
		Get(node interface{}) (value interface{}, has bool)
		Remove(node interface{})
		// Cascade 多次hash排序后的hash数组
		Cascade(node interface{}, opts ...ConsistentAddWith) []uint64
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

	return &consistent{
		opt:     dft,
		vtrNum:  dft.vtr,
		vtrKeys: make([]uint64, 0),
		vtrRing: make(map[uint64][]interface{}),
		nodes:   make(map[string]struct{}),
	}
}

func (c *consistent) Cascade(node interface{}, opts ...ConsistentAddWith) []uint64 {
	dft := &optionAdd{vtr: c.vtrNum, weight: 0}
	for i := range opts {
		opts[i](dft)
	}

	vtr := c.vtrNum
	if dft.weight > 0 {
		vtr = c.vtrNum * dft.weight / maxWeight
	}

	if vtr > c.vtrNum {
		vtr = c.vtrNum
	}

	nodeExpr := c.marshal(node)
	c.lock.Lock()
	defer c.lock.Unlock()

	keys := make([]uint64, 0)
	for i := int64(0); i < vtr; i++ {
		hashV := c.opt.hash([]byte(nodeExpr + strconv.Itoa(int(i))))
		keys = append(keys, hashV)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return keys
}

func (c *consistent) Add(node interface{}, opts ...ConsistentAddWith) {
	dft := &optionAdd{vtr: c.vtrNum, weight: 0}
	for i := range opts {
		opts[i](dft)
	}

	vtr := c.vtrNum
	if dft.weight > 0 {
		vtr = c.vtrNum * dft.weight / maxWeight
	}

	if vtr > c.vtrNum {
		vtr = c.vtrNum
	}

	nodeExpr := c.marshal(node)
	c.lock.Lock()
	defer c.lock.Unlock()
	c.addNode(nodeExpr)

	for i := int64(0); i < vtr; i++ {
		hashV := c.opt.hash([]byte(nodeExpr + strconv.Itoa(int(i))))
		c.vtrKeys = append(c.vtrKeys, hashV)
		c.vtrRing[hashV] = append(c.vtrRing[hashV], node)
	}

	sort.Slice(c.vtrKeys, func(i, j int) bool {
		return c.vtrKeys[i] < c.vtrKeys[j]
	})
}

func (c *consistent) Get(node interface{}) (value interface{}, has bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if len(c.vtrRing) == 0 {
		return nil, false
	}

	nodeExpr := c.marshal(node)
	hashV := c.opt.hash([]byte(nodeExpr))
	index := sort.Search(len(c.vtrKeys), func(i int) bool {
		return c.vtrKeys[i] >= hashV
	}) % len(c.vtrKeys)

	nodes := c.vtrRing[c.vtrKeys[index]]
	switch len(nodes) {
	case 0:
		return nil, false
	case 1:
		return nodes[0], true
	default:
		index := c.opt.hash([]byte(c.innerMarshal(node)))
		pos := int(index % uint64(len(nodes)))
		return nodes[pos], true
	}
}

func (c *consistent) Remove(node interface{}) {
	nodeExpr := c.marshal(node)

	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.nodes[nodeExpr]; !ok {
		return
	}

	for i := int64(0); i < c.vtrNum; i++ {
		hashV := c.opt.hash([]byte(nodeExpr + strconv.Itoa(int(i))))
		index := sort.Search(len(c.vtrKeys), func(i int) bool {
			return c.vtrKeys[i] >= hashV
		})
		if index < len(c.vtrKeys) && c.vtrKeys[index] == hashV {
			c.vtrKeys = append(c.vtrKeys[:index], c.vtrKeys[index+1:]...)
		}

		if _, ok := c.vtrRing[hashV]; ok {
			newNodes := c.vtrRing[hashV][:0]
			for _, node := range c.vtrRing[hashV] {
				if c.marshal(node) != nodeExpr {
					newNodes = append(newNodes, node)
				}
			}
			if len(newNodes) > 0 {
				c.vtrRing[hashV] = newNodes
			} else {
				delete(c.vtrRing, hashV)
			}
		}
	}

	delete(c.nodes, nodeExpr)
}

func (c *consistent) addNode(key string) {
	c.nodes[key] = struct{}{}
}

func (c *consistent) marshal(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

func (c *consistent) innerMarshal(node interface{}) string {
	return fmt.Sprintf("%d:%v", prime, node)
}
