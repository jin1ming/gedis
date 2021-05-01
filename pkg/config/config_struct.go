package config

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
	SlaveReadOnly	bool		`yaml:"slaveReadOnly"`
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
