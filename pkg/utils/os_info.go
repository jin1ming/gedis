package utils

import (
	"log"
	"os"
	"os/user"
)

var HOMEDIR string

func init() {
	HOMEDIR = os.Getenv("HOME")
	if HOMEDIR == "" {
		if u, err := user.Current(); err == nil {
			HOMEDIR = u.HomeDir
			log.Println(HOMEDIR)
		}
	}
}

func GetHomeDir() string {
	return HOMEDIR
}
