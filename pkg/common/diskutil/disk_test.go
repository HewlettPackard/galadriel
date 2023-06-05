package diskutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "galadriel-test")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(tempDir))
	})

	tests := []struct {
		name            string
		data            []byte
		atomicWriteFunc func(string, []byte) error
		expectMode      os.FileMode
	}{
		{
			name:            "basic - AtomicWritePrivateFile",
			data:            []byte("Hello, World"),
			atomicWriteFunc: AtomicWritePrivateFile,
			expectMode:      0600,
		},
		{
			name:            "empty - AtomicWritePrivateFile",
			data:            []byte{},
			atomicWriteFunc: AtomicWritePrivateFile,
			expectMode:      0600,
		},
		{
			name:            "binary - AtomicWritePrivateFile",
			data:            []byte{0xFF, 0, 0xFF, 0x3D, 0xD8, 0xA9, 0xDC, 0xF0, 0x9F, 0x92, 0xA9},
			atomicWriteFunc: AtomicWritePrivateFile,
			expectMode:      0600,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			file := filepath.Join(tempDir, "file")
			err := tt.atomicWriteFunc(file, tt.data)
			require.NoError(t, err)

			info, err := os.Stat(file)
			require.NoError(t, err)
			require.EqualValues(t, tt.expectMode, info.Mode())

			content, err := os.ReadFile(file)
			require.NoError(t, err)
			require.Equal(t, tt.data, content)

			require.NoError(t, os.Remove(file))
		})
	}
}
