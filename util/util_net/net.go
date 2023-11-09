package util_net

import (
	"fmt"
	"net"
	"time"
)

func CheckTcpService(addrArr []string) (allOk bool, e error) {
	len := len(addrArr)
	for i := 0; i < len; i++ {
		addr := addrArr[i]
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err != nil {
			fmt.Printf("addr:  %s ok NOT!!! \n", addr)
			return false, nil
		}
		defer conn.Close()

		fmt.Printf("addr:  %s ok\n", addr)
	}

	return true, nil
}
