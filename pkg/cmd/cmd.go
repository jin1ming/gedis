package cmd

import (
	"golang.org/x/tools/go/ssa/interp/testdata/src/strings"
	"os"
)

func main() {
	configFile := ""
	if len(os.Args) > 0 && strings.Contains(os.Args[0], ".conf") {
		configFile = os.Args[0]
	} else {
		configFile = "redis.conf"
	}

	// TODO: config解析

	// TODO: 相关监听服务的启动
}