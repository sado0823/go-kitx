package internal

//func Test_1(t *testing.T) {
//	cgroup, err := currentCgroup()
//	t.Log("Test_1:", err)
//	if err != nil {
//		return
//	}
//	us, err := cgroup.cpuPeriodUs()
//	t.Log("us:", us, "err:", err)
//	quotaUS, err := cgroup.cpuQuotaUs()
//	t.Log("quotaUS:", quotaUS, "err:", err)
//	cpus, err := cgroup.cpus()
//	t.Log("cpus:", cpus, "err:", err)
//	allCpus, err := cgroup.usageAllCpus()
//	t.Log("allCpus:", allCpus, "err:", err)
//
//	tc := time.NewTicker(time.Second)
//	go func() {
//		for {
//			select {
//			case <-tc.C:
//				cpu := RefreshCpu()
//				t.Log("now cpu:", cpu)
//			default:
//				continue
//			}
//		}
//	}()
//	for {
//
//	}
//}
