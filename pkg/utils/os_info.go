package utils

import (
	"os"
	"os/user"
)

var HOMEDIR string

func init() {
	home := os.Getenv("HOME")
	if home == "" {
		if u, err := user.Current(); err == nil {
			HOMEDIR = u.HomeDir
		}
	}
}

func GetHomeDir() string {
	return HOMEDIR
}
