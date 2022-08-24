package config

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

var newFn = New

// LoadFromDisk reads the configuration from the given file path.
func LoadFromDisk(path string) (*HarvesterConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("unable to open configuration file: %v", err))
	}
	defer file.Close()

	return newFn(file)
}
