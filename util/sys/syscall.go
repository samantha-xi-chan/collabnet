package sys

import (
	"fmt"
	"syscall"
)

func PrintLoadAverage() {
	var info syscall.Sysinfo_t

	if err := syscall.Sysinfo(&info); err != nil {
		fmt.Println("Error getting system information:", err)
		return
	}

	// 负载平均值是最近1分钟、5分钟和15分钟的平均值
	load1 := float64(info.Loads[0]) / 65536.0
	load5 := float64(info.Loads[1]) / 65536.0
	load15 := float64(info.Loads[2]) / 65536.0

	fmt.Printf("Load Average: %.2f %.2f %.2f\n", load1, load5, load15)
}
