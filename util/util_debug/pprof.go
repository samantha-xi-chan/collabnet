package util_debug

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
)

func InitPProf(addr string) {
	// 开启pprof，监听请求 http://localhost:6060/debug/pprof/
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("start pprof failed on %s\n", addr)
	}
}
