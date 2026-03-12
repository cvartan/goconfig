package utils

import "os"

func GetWD() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}

	panic("[config:getWd:01] Can't get working directory")
}
