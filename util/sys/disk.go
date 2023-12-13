package sys

import (
	"fmt"
	"syscall"
)

func PrintDisk() {
	path := "/"

	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	totalSpace := stat.Blocks * uint64(stat.Bsize)
	availableSpace := stat.Bavail * uint64(stat.Bsize)

	fmt.Printf("All   Space: %d bytes\n", totalSpace)
	fmt.Printf("Avail Space: %d bytes\n", availableSpace)
}
