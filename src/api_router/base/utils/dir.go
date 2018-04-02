package utils

import (
	"os"
	"path/filepath"
	"strings"
	l4g "github.com/alecthomas/log4go"
)

func GetRunDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		l4g.Error(err)
		os.Exit(1)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func GetCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		l4g.Error(err)
		os.Exit(1)
	}
	return strings.Replace(dir, "\\", "/", -1)
}


func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}