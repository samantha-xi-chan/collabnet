package util_os

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
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
