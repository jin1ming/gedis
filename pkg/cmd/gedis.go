package main

import (
	"context"
	"github.com/jin1ming/Gedis/pkg/cli"
	"github.com/jin1ming/Gedis/pkg/config"
	"github.com/jin1ming/Gedis/pkg/ps"
	"github.com/jin1ming/Gedis/pkg/server"
	"os"
	"runtime"
	"strings"
)

var buildVersion = "0.0.1"

func main() {
	runtime.LockOSThread()

	configFile := "../config/gedis.yaml"
	if len(os.Args) > 0 && strings.Contains(os.Args[0], ".yaml") {
		configFile = os.Args[0]
	}

	// config
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
	cfg := config.GetConfig()

	ctx, cancel := context.WithCancel(context.Background())

	if cfg.Append.AppendOnly {
		go ps.AOFService{}.Start(ctx)
	}

	s := server.New()
	s.Start()

	cancel() // TODO
}
