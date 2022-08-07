//go:build !linux
// +build !linux

package internal

// RefreshCpu return 0 without running on linux
func RefreshCpu() uint64 {
	return 0
}
