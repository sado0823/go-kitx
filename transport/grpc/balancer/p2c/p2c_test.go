package p2c

import (
	"context"
	"fmt"
	"github.com/sado0823/go-kitx/pkg/stringx"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/status"
	"math"
	"runtime"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestP2cPicker_PickNil(t *testing.T) {
	builder := new(p2cPickBuilder)
	picker := builder.Build(base.PickerBuildInfo{})
	pick, err := picker.Pick(balancer.PickInfo{
		FullMethodName: "/",
		Ctx:            context.Background(),
	})
	fmt.Println(pick)
	assert.NotNil(t, err)
}

func TestP2cPicker_Pick(t *testing.T) {
	cases := []struct {
		name      string
		choices   int
		err       error
		threshold float64
	}{
		{
			name: "empty", choices: 0, err: balancer.ErrNoSubConnAvailable,
		},
		{
			name: "single", choices: 1, threshold: 0.9,
		},
		{
			name: "two", choices: 2, threshold: 0.5,
		},
		{
			name: "multiple", choices: 100, threshold: 0.95,
		},
	}

	for _, caseV := range cases {
		caseV := caseV
		t.Run(caseV.name, func(t *testing.T) {
			t.Parallel()

			const total = 10000
			builder := new(p2cPickBuilder)

			ready := make(map[balancer.SubConn]base.SubConnInfo)
			for i := 0; i < caseV.choices; i++ {
				ready[mockClientConn{id: stringx.Rand(6)}] = base.SubConnInfo{
					Address: resolver.Address{
						Addr: strconv.Itoa(i),
					},
				}
			}

			picker := builder.Build(base.PickerBuildInfo{ReadySCs: ready})

			var wg sync.WaitGroup
			wg.Add(total)

			for i := 0; i < total; i++ {
				result, err := picker.Pick(balancer.PickInfo{
					FullMethodName: "/",
					Ctx:            context.Background(),
				})
				assert.Equal(t, caseV.err, err)

				if caseV.err != nil {
					return
				}

				if i%100 == 0 {
					err = status.Error(codes.DeadlineExceeded, "deadline")
				}

				go func() {
					runtime.Gosched()
					result.Done(balancer.DoneInfo{
						Err: err,
					})
					wg.Done()
				}()
			}

			wg.Wait()

			dist := make(map[string]int)
			conns := picker.(*p2cPicker).conns
			for _, conn := range conns {
				dist[conn.addr.Addr] = int(conn.requests)
			}

			entropy := calcEntropy(dist)
			fmt.Printf("entropy:%v, threshold:%v \n", entropy, caseV.threshold)
			assert.True(t, entropy > caseV.threshold, fmt.Sprintf("entropy is %f, less than %f",
				entropy, caseV.threshold))

		})
	}

}

type mockClientConn struct {
	// add random string member to avoid map key equality.
	id string
}

func (m mockClientConn) UpdateAddresses(addresses []resolver.Address) {
}

func (m mockClientConn) Connect() {
}

// calcEntropy calculates the entropy of m.
func calcEntropy(m map[string]int) float64 {
	if len(m) == 0 || len(m) == 1 {
		return 1
	}

	const epsilon = 1e-6

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
