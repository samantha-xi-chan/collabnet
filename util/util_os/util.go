package util_os

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func GetMaxOpenFiles() (int, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.Command("bash", "-c", "ulimit -n")
	case "windows":
		// Unfortunately, Windows does not have a direct equivalent to ulimit.
		// You may need to use external tools or libraries specific to Windows.
		return 0, fmt.Errorf("unsupported on Windows")
	default:
		return 0, fmt.Errorf("unsupported operating system")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("failed to get max open files: %v", err)
	}

	// Convert the output to an integer
	maxOpenFilesStr := strings.TrimSpace(string(output))
	maxOpenFiles, err := strconv.Atoi(maxOpenFilesStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse max open files: %v", err)
	}

	return maxOpenFiles, nil
}

func PrintCpuMemUsage() {
	cpuUsage, err := cpu.Percent(time.Second, false)
	if err != nil {
		fmt.Println("Error getting CPU usage:", err)
	} else {
		fmt.Printf("CPU Usage: %.2f%%\n", cpuUsage[0])
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println("Error getting memory usage:", err)
	} else {
		fmt.Printf("Total: %v, Free: %v, Used: %v, UsedPercent: %.2f%%\n", memInfo.Total, memInfo.Free, memInfo.Used, memInfo.UsedPercent)
	}
}

/*
func main() {
	maxOpenFiles, err := GetMaxOpenFiles()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println("Max open files:", maxOpenFiles)
}
*/
