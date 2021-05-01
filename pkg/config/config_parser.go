package config

import (
	"gopkg.in/yaml.v2"
	"os"
)

type BaseConfig struct {
	Bind			string		`yaml:"bind"`
	Port			int			`yaml:"port"`
	MaxClients		int			`yaml:"maxClients"`
	PidFile			string		`yaml:"pidFile"`
	LogFile			string		`yaml:"logFile"`
	LogLevel		string		`yaml:"logLevel"`
	DbFileName		string		`yaml:"dbFileName"`
	DbFileDir		string		`yaml:"dbFileDIr"`
	MaxMemory		int			`yaml:"maxMemory"`
}

type MSMode struct {
	SlaveOf			string		`yaml:"slaveOf"`
	MasterAuth		string		`yaml:"masterAuth"`
	SlaveReadOnly	string		`yaml:"slaveReadOnly"`
}

type AppendMode struct {
	AppendOnly		bool		`yaml:"appendOnly"`
	AppendFileName	string		`yaml:"appendFileName"`
	AppendFsync		string		`yaml:"appendFsync"`
}

type Config struct {
	Base	BaseConfig 	`yaml:"base"`
	MS		MSMode		`yaml:"masterSlaveMode"`
	Append	AppendMode	`yaml:"appendMode"`
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}
	if file, err := os.Open(path); err != nil {
		return nil, err
	} else {
		err = yaml.NewDecoder(file).Decode(cfg)
		if err != nil {
			return nil, err
		}
	}
	return cfg, nil
}