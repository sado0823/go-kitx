package internal

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	cpuTicks  = 100
	cpuFields = 8
)

var (
	preSystem uint64
	preTotal  uint64
	quota     float64
	cores     uint64
)

func init() {
	cpus, err := cpuSets()
	if err != nil {
		logger.Printf("cpuSets err:%+v", err)
		return
	}

	cores = uint64(len(cpus))
	sets, err := cpuSets()
	if err != nil {
		logger.Printf("cpuSets err:%+v", err)
		return
	}
	quota = float64(len(sets))

	cq, err := cpuQuota()
	if err == nil {
		if cq != -1 {
			period, err := cpuPeriod()
			if err != nil {
				logger.Printf("cpuPeriod err:%+v", err)
				return
			}

			limit := float64(cq) / float64(period)
			if limit < quota {
				quota = limit
			}
		}
	}

	preSystem, err = systemCpuUsage()
	if err != nil {
		logger.Printf("systemCpuUsage err:%+v", err)
		return
	}

	preTotal, err = totalCpuUsage()
	if err != nil {
		logger.Printf("totalCpuUsage err:%+v", err)
		return
	}
}

func RefreshCpu() uint64 {
	total, err := totalCpuUsage()
	if err != nil {
		return 0
	}
	system, err := systemCpuUsage()
	if err != nil {
		return 0
	}

	var usage uint64
	cpuDelta := total - preTotal
	systemDelta := system - preSystem
	if cpuDelta > 0 && systemDelta > 0 {
		usage = uint64(float64(cpuDelta*cores*1e3) / (float64(systemDelta) * quota))
	}
	preSystem = system
	preTotal = total

	return usage
}

func systemCpuUsage() (uint64, error) {
	/**
	//CPU指标：user，nice, system, idle, iowait, irq, softirq
	cpu  23149 0 49899 9780030 443 0 1219 0 0 0
	cpu0 11685 0 24546 4890485 241 0 851 0 0 0
	cpu1 11463 0 25352 4889545 201 0 367 0 0 0
	*/
	lines, err := readLines("/proc/stat")
	if err != nil {
		return 0, err
	}

	for _, line := range lines {
		cols := strings.Fields(line)
		if cols[0] == "cpu" {
			if len(cols) < cpuFields {
				return 0, fmt.Errorf("bad format of cpu stats")
			}
		}

		var totalClockTicks uint64
		for _, v := range cols[1:cpuFields] {
			u, err := parseUint(v)
			if err != nil {
				return 0, err
			}
			totalClockTicks += u
		}
		return (totalClockTicks * uint64(time.Second)) / cpuTicks, nil

	}
	return 0, errors.New("bad cpu stats format")
}

func cpuSets() ([]uint64, error) {
	c, err := currentCgroup()
	if err != nil {
		return nil, err
	}
	return c.cpus()
}

func cpuQuota() (int64, error) {
	cg, err := currentCgroup()
	if err != nil {
		return 0, err
	}

	return cg.cpuQuotaUs()
}

func cpuPeriod() (uint64, error) {
	cg, err := currentCgroup()
	if err != nil {
		return 0, err
	}

	return cg.cpuPeriodUs()
}

func totalCpuUsage() (usage uint64, err error) {
	var cg cgroup
	if cg, err = currentCgroup(); err != nil {
		return
	}

	return cg.usageAllCpus()
}