package main

import (
	"context"
	"github.com/jin1ming/Gedis/pkg/cli"
	"github.com/jin1ming/Gedis/pkg/config"
	"github.com/jin1ming/Gedis/pkg/ps"
	"github.com/jin1ming/Gedis/pkg/server"
	"github.com/tidwall/redcon"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
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

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		info := <-sigs
		log.Println("Signal:", info)
		cancel()
	}()

	var aofBuffer chan redcon.Command
	if cfg.Append.AppendOnly {
		aofService := ps.NewAOFService()
		aofBuffer = aofService.ChBuffer
		aofService.LoadLocalData()
		go aofService.Start(ctx)
	}

	s := server.New(aofBuffer)
	s.Start(ctx)

}
