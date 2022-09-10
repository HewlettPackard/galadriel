package common

import (
	"fmt"
	"net"
	"path/filepath"
)

func GetAbsoluteUDSPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("unable to get absolute path of %s: %v", path, err)
	}

	c, err := net.Dial("unix", absPath)
	if err != nil {
		return "", fmt.Errorf("unable to dial UDS path %s: %v", absPath, err)
	}
	defer c.Close()

	absPath = "unix://" + path

	return absPath, nil
}
