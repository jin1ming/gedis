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
	DefaultConfigDIr = ".gedis"
)

var (
	initConfigDir = new(sync.Once)
	configDir	string
)

func setConfigDir() {
	if configDir != "" {
		return
	}
	configDir = os.Getenv("GEDIS_CONFIG")
	if configDir == "" {
		configDir = filepath.Join(utils.GetHomeDir(), DefaultConfigDIr)
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

func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configDir = Dir()
		configPath = filepath.Join(configDir, DefaultConfigName)
	}
	configPath = filepath.Clean(configPath)

	cfg := &Config{}
	if file, err := os.Open(configPath); err != nil {
		return nil, err
	} else {
		err = yaml.NewDecoder(file).Decode(cfg)
		if err != nil {
			return nil, err
		}
	}
	return cfg, nil
}
