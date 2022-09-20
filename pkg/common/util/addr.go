package util

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

// GetUnixAddrWithAbsPath converts a string path to a net.UnixAddr
// making the path absolute.
func GetUnixAddrWithAbsPath(path string) (*net.UnixAddr, error) {
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for socket path: %w", err)
	}

	return &net.UnixAddr{
		Name: pathAbs,
		Net:  "unix",
	}, nil
}

// PrepareLocalAddr creates the folders in the path for the localAddr
func PrepareLocalAddr(localAddr net.Addr) error {
	if err := os.MkdirAll(filepath.Dir(localAddr.String()), 0750); err != nil {
		return fmt.Errorf("unable to create socket directory: %w", err)
	}

	return nil
}
