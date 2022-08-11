package sysx

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sado0823/go-kitx/pkg/sysx/internal"
)

var (
	logger = log.New(os.Stdout, fmt.Sprintf("[DEBUG][pkg=sysx][%s] ", time.Now().Format(time.StampMilli)), log.Lshortfile)
)

func init() {
	//logger.SetFlags(0)
	//logger.SetOutput(io.Discard)
}

const (
	cpuRefreshInterval = time.Millisecond * 250
	allRefreshInterval = time.Minute * 1
	beta               = 0.95
)

var cpuUsage int64

func init() {
	go func() {
		cpuTicker := time.NewTicker(cpuRefreshInterval)
		defer cpuTicker.Stop()
		allTicker := time.NewTicker(allRefreshInterval)
		defer allTicker.Stop()

		for {
			select {
			case <-cpuTicker.C:
				go func() {
					defer func() {
						if e := recover(); e != nil {
							logger.Printf("cpuTicker panic, stack:%s", printStack())
						}
						current := internal.RefreshCpu()
						pre := atomic.LoadInt64(&cpuUsage)
						// cpu = cpuᵗ⁻¹ * beta + cpuᵗ * (1 - beta)
						usage := int64(float64(pre)*beta + float64(current)*(1-beta))
						atomic.StoreInt64(&cpuUsage, usage)
					}()
				}()
			case <-allTicker.C:
				printUsage()
			}
		}

	}()

}

func CpuUsage() int64 {
	return atomic.LoadInt64(&cpuUsage)
}

func bToMb(b uint64) float32 {
	return float32(b) / 1024 / 1024
}

func printUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	logger.Printf("CPU: %dm, MEMORY: Alloc=%.1fMi, TotalAlloc=%.1fMi, Sys=%.1fMi, NumGC=%d",
		CpuUsage(), bToMb(m.Alloc), bToMb(m.TotalAlloc), bToMb(m.Sys), m.NumGC)
}

func printStack() []string {
	//buf := make([]byte, 64<<10)
	buf := make([]byte, 1<<10*64)
	n := runtime.Stack(buf, false)
	buf = buf[:n]
	stack := strings.Split(fmt.Sprintln(string(buf[:n])), "\n")
	return stack
}
