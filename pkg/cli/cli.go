package cli

import (
	"github.com/jin1ming/Gedis/pkg/client"
	"github.com/jin1ming/Gedis/pkg/config"
)

type GedisCli struct {
	config     *config.Config
	client     []*client.Client
	ServerInfo *ServerInfo
	clientInfo []*ClientInfo
}

func (cli *GedisCli) SetConfig(configPath string) error {
	if cli.config == nil {
		configTmp, err := config.LoadConfig(configPath)
		if err != nil {
			return err
		}
		cli.config = configTmp
	}
	return nil
}

func (cli *GedisCli) GetConfig() *config.Config {
	return cli.config
}

type ServerInfo struct {
	OSType          string
	BuildVersion	string
}

type ClientInfo struct {
	OSType          string
	DefaultVersion  string
}