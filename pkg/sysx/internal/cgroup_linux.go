package internal

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
	"io"

	"golang.org/x/sys/unix"
)

var (
	logger = log.New(os.Stdout, fmt.Sprintf("[DEBUG][pkg=sysx/internal][%s] ", time.Now().Format(time.StampMilli)), log.Lshortfile)
)

func init() {
	logger.SetFlags(0)
	logger.SetOutput(io.Discard)
}

const (
	cgroupDir         = "/sys/fs/cgroup"
	cgroupCpuStatFile = cgroupDir + "/cpu.stat"
	cgroupCpusetFile  = cgroupDir + "/cpuset.cpus.effective"
)

var (
	isUnifiedOnce sync.Once
	// is cgroup v2 unified mode
	isUnified bool

	nsOnce sync.Once
	// is in user namespace
	inUserNS bool
)

type cgroup interface {
	// cfs_quota_us表示Cgroup可以使用的cpu的带宽，单位为微秒。
	// cfs_quota_us为-1，表示使用的CPU不受cgroup限制。
	// cfs_quota_us的最小值为1ms(1000)，最大值为1s。
	cpuQuotaUs() (int64, error)
	// cfs_period_us表示一个cpu带宽，单位为微秒。系统总CPU带宽： cpu核心数 * cfs_period_us
	cpuPeriodUs() (uint64, error)
	cpus() ([]uint64, error)
	usageAllCpus() (uint64, error)
}

func currentCgroup() (cgroup cgroup, err error) {
	if isCgroupV2Unified() {
		logger.Println("using cgroup v2")
		return newCgroupV2()
	}
	logger.Println("using cgroup v1")
	return newCgroupV1()
}

func isCgroupV2Unified() bool {
	isUnifiedOnce.Do(func() {
		var st unix.Statfs_t
		err := unix.Statfs(cgroupDir, &st)
		if err != nil {
			if os.IsNotExist(err) && runningInUserNS() {
				isUnified = false
				return
			}
			panic(fmt.Sprintf("cannot statfs cgroup root: %s", err))
		}
		isUnified = st.Type == unix.CGROUP2_SUPER_MAGIC
	})

	return isUnified
}

// running in user namespace or not
func runningInUserNS() bool {
	nsOnce.Do(func() {
		// https://man7.org/linux/man-pages/man7/user_namespaces.7.html
		// ID-inside-ns   ID-outside-ns   length
		file, err := os.Open("/proc/self/uid_map")
		logger.Printf("/proc/self/uid_map open err: %+v", err)
		if err != nil {
			// can only find this file in user namespace
			return
		}
		defer file.Close()
		line, _, err := bufio.NewReader(file).ReadLine()
		if err != nil {
			return
		}
		lineStr := string(line)
		var inside, outside, length int64
		fmt.Sscanf(lineStr, "%d %d %d", &inside, &outside, &length)
		logger.Printf("/proc/self/uid_map:%s, inside:%d, outside:%d, length:%d", lineStr, &inside, &outside, &length)

		if inside == 0 && outside == 0 && length == 4294967295 {
			return
		}
		inUserNS = true
	})

	return inUserNS
}

func newCgroupV1() (cgroup, error) {
	cgroupFile := fmt.Sprintf("/proc/%d/cgroup", os.Getpid())
	lines, err := readLines(cgroupFile)
	if err != nil {
		return nil, err
	}

	/**
	// 得到类似如下信息 (我这里读的是某个docker进程的数据)
	//11:cpuset:/docker/290247cde1fff59d5322068be83a7c7629f4454ac0960a89e6856ea041970b30
	//10:memory:/docker/290247cde1fff59d5322068be83a7c7629f4454ac0960a89e6856ea041970b30
	//9:devices:/docker/290247cde1fff59d5322068be83a7c7629f4454ac0960a89e6856ea041970b30
	//8:blkio:/docker/290247cde1fff59d5322068be83a7c7629f4454ac0960a89e6856ea041970b30
	//7:hugetlb:/docker/290247cde1fff59d5322068be83a7c7629f4454ac0960a89e6856ea041970b30
	//6:perf_event:/docker/290247cde1fff59d5322068be83a7c7629f4454ac0960a89e6856ea041970b30
	//5:freezer:/docker/290247cde1fff59d5322068be83a7c7629f4454ac0960a89e6856ea041970b30
	//4:net_cls,net_prio:/docker/290247cde1fff59d5322068be83a7c7629f4454ac0960a89e6856ea041970b30
	//3:pids:/docker/290247cde1fff59d5322068be83a7c7629f4454ac0960a89e6856ea041970b30
	//2:cpu,cpuacct:/docker/290247cde1fff59d5322068be83a7c7629f4454ac0960a89e6856ea041970b30
	//1:name=systemd:/docker/290247cde1fff59d5322068be83a7c7629f4454ac0960a89e6856ea041970b30
	*/
	cgroups := make(map[string]string)
	for _, line := range lines {
		cols := strings.Split(line, ":")
		if len(cols) != 3 {
			return nil, fmt.Errorf("invalid cgroup v1 line:%s", line)
		}
		subsys := cols[1]
		// only read cpu
		if !strings.HasPrefix(subsys, "cpu") {
			continue
		}

		// https://man7.org/linux/man-pages/man7/cgroups.7.html
		fields := strings.Split(subsys, ",")
		for _, field := range fields {
			cgroups[field] = path.Join(cgroupDir, field)
		}
	}

	return &cgroupV1{cgroups: cgroups}, nil
}

// https://docs.kernel.org/admin-guide/cgroup-v1/index.html
type cgroupV1 struct {
	// map[subsys.field]path
	cgroups map[string]string
}

func (c cgroupV1) cpuQuotaUs() (int64, error) {
	text, err := readText(path.Join(c.cgroups["cpu"], "cpu.cfs_quota_us"))
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(text, 10, 64)
}

func (c cgroupV1) cpuPeriodUs() (uint64, error) {
	text, err := readText(path.Join(c.cgroups["cpu"], "cpu.cfs_period_us"))
	if err != nil {
		return 0, err
	}
	return parseUint(text)
}

func (c cgroupV1) cpus() ([]uint64, error) {
	data, err := readText(path.Join(c.cgroups["cpuset"], "cpuset.cpus"))
	if err != nil {
		return nil, err
	}
	return parseUints(data)
}

func (c cgroupV1) usageAllCpus() (uint64, error) {
	data, err := readText(path.Join(c.cgroups["cpuacct"], "cpuacct.usage"))
	if err != nil {
		return 0, err
	}
	return parseUint(data)
}

func newCgroupV2() (cgroup, error) {
	lines, err := readLines(cgroupCpuStatFile)
	if err != nil {
		return nil, err
	}

	/**
	// lines like this:
	//usage_usec 27650012  - 占用cpu总时间
	//user_usec 12736431   - 用户态占用时间
	//system_usec 14913580 - 内核态占用时间
	//nr_periods 0         - 周期计数
	//nr_throttled 0       - 周期内的限制计数
	//throttled_usec 0     - 限制执行的时间
	*/
	cgroups := make(map[string]string)
	for _, line := range lines {
		cols := strings.Fields(line)
		if len(cols) != 2 {
			return nil, fmt.Errorf("invalid cgroup v2 line:%s", line)
		}
		cgroups[cols[0]] = cols[1]
	}
	return &cgroupV2{cgroups: cgroups}, nil
}

// https://docs.kernel.org/admin-guide/cgroup-v2.html
type cgroupV2 struct {
	// map[subsys.field]value
	cgroups map[string]string
}

func (c cgroupV2) cpuQuotaUs() (int64, error) {
	data, err := readText(path.Join(cgroupDir, "cpu.cfs_quota_us"))
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(data, 10, 64)
}

func (c cgroupV2) cpuPeriodUs() (uint64, error) {
	data, err := readText(path.Join(cgroupDir, "cpu.cfs_period_us"))
	if err != nil {
		return 0, err
	}

	return parseUint(data)
}

func (c cgroupV2) cpus() ([]uint64, error) {
	// 显示当前cgroup真实可用的cpu列表
	// lines like: 0-1
	data, err := readText(cgroupCpusetFile)
	if err != nil {
		return nil, err
	}

	return parseUints(data)
}

func (c cgroupV2) usageAllCpus() (uint64, error) {
	usec, err := parseUint(c.cgroups["usage_usec"])
	if err != nil {
		return 0, err
	}

	return usec * uint64(time.Microsecond), nil
}
