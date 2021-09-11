package utils

import (
	"os"
	"os/user"
)

func HomeDir() string {
	// TODO: windows还不确定是否可以
	return "HOME"
}

func GetHomeDir() string {
	home := os.Getenv(HomeDir())
	if home == "" {
		if u, err := user.Current(); err == nil {
			return u.HomeDir
		}
	}
	return home
}
