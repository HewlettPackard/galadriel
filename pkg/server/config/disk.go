package config

import (
	"fmt"
	"os"
)

var newFn = New

// LoadFromDisk reads the configuration from the given file path.
func LoadFromDisk(path string) (*Server, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open configuration file: %v", err)
	}
	defer file.Close()

	return newFn(file)
}
