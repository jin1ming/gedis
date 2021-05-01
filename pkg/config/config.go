package config

import (
	"github.com/jin1ming/Gedis/pkg/utils"
	"os"
	"path/filepath"
	"sync"
)

const (
	DefaultConfigName = "gedis.conf"
	DefaultConfigDIr = ".gedis"
)

var (
	initConfigDIr = new(sync.Once)
	configDir	string
	homeDir		string
)

func setConfigDir() {
	if configDir != "" {
		return
	}
	configDir = os.Getenv("GEDIS_CONFIg")
	if configDir == "" {
		configDir = filepath.Join(utils.GetHomeDir(), DefaultConfigDIr)
	}
}

func resetConfigDir() {
	configDir = ""
	initConfigDIr = new(sync.Once)
}

func Dir() string {
	initConfigDIr.Do(setConfigDir)
	return configDir
}

func SetDir(dir string) {
	configDir = filepath.Clean(dir)
}

func Load(configDir string) () {

}