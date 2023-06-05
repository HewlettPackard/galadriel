package diskutil

import (
	"os"
	"path/filepath"
)

// Define the file mode for private files.
const (
	fileModePrivate = 0600
)

// AtomicWritePrivateFile writes data to a file atomically.
// The file is created with private permissions (0600).
func AtomicWritePrivateFile(path string, data []byte) error {
	return atomicWrite(path, data, fileModePrivate)
}

func atomicWrite(path string, data []byte, mode os.FileMode) error {
	tmpPath := path + ".tmp"

	// Attempt to write to temporary file
	if err := write(tmpPath, data, mode); err != nil {
		return err
	}

	// If write is successful, rename temp file to actual path
	return rename(tmpPath, path)
}

func write(path string, data []byte, mode os.FileMode) error {
	// Attempt to open file, deferring close until function finishes
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	// Sync file contents to disk
	return file.Sync()
}

func rename(oldPath, newPath string) error {
	// Attempt to rename file
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}

	// Open containing directory and defer close until function finishes
	dir, err := os.Open(filepath.Dir(newPath))
	if err != nil {
		return err
	}
	defer dir.Close()

	// Sync directory changes to disk
	return dir.Sync()
}
