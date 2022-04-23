package utils

import (
	"log"
	"os"
	"testing"
)

func TestOsInfo(t *testing.T) {
	home := GetHomeDir()
	s, err := os.Stat(home)
	if err != nil || !s.IsDir() {
		log.Fatalln("home dir can't find.")
	}
}
