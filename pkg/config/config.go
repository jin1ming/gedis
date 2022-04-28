package config

import (
	"github.com/jin1ming/Gedis/pkg/utils"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"sync"
)

const (
	DefaultConfigName = "gedis.yaml"
	DefaultConfigDIr  = ".gedis"
)

var (
	initConfigDir = new(sync.Once)
	configDir     string
	serverConfig  *Config
)

func setConfigDir() {
	if configDir != "" {
		return
	}
	configDir = os.Getenv("GEDIS_CONFIG")
	if configDir == "" {
		configDir = filepath.Join(utils.GetHomeDir(), DefaultConfigDIr)
	}
	_, err := os.Stat(configDir)
	if err != nil && os.IsNotExist(err) {
		os.Mkdir(configDir, os.ModeDir)
	}
}

func resetConfigDir() {
	configDir = ""
	initConfigDir = new(sync.Once)
}

func Dir() string {
	initConfigDir.Do(setConfigDir)
	return configDir
}

func LoadConfig(configPath string) error {
	if configPath == "" {
		configDir = Dir()
		configPath = filepath.Join(configDir, DefaultConfigName)
	}
	configPath = filepath.Clean(configPath)

	cfg := &Config{}
	if file, err := os.Open(configPath); err != nil {
		return err
	} else {
		err = yaml.NewDecoder(file).Decode(cfg)
		if err != nil {
			return err
		}
	}
	serverConfig = cfg
	return nil
}

func GetConfig() *Config {
	return serverConfig
}
