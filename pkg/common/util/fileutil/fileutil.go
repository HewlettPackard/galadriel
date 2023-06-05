package fileutil

import (
	"errors"
	"os"
)

// CreateDirIfNotExist creates a directory at the specified path if it doesn't exist
func CreateDirIfNotExist(dirPath string) error {
	if dirPath == "" {
		return errors.New("directory path is required")
	}

	// Check if the directory already exists
	_, err := os.Stat(dirPath)
	if err == nil {
		// Directory already exists
		return nil
	}

	// Create the directory
	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
