package main

import (
	"github.com/jin1ming/Gedis/pkg/cli"
	"github.com/jin1ming/Gedis/pkg/server"
	"os"
	"runtime"
	"strings"
)

var buildVersion = "0.0.1"

func main() {
	configFile := "../config/gedis.yaml"
	if len(os.Args) > 0 && strings.Contains(os.Args[0], ".yaml") {
		configFile = os.Args[0]
	}

	// config解析
	gedis := &cli.GedisCli{
		ServerInfo: &cli.ServerInfo{
			OSType:       runtime.GOOS,
			BuildVersion: buildVersion,
		},
	}
	err := gedis.SetConfig(configFile)
	if err != nil {
		panic(err)
	}
	//fmt.Println(config.GetConfig())

	// TODO: 相关监听服务的启动
	s := server.New()
	s.Start()
}