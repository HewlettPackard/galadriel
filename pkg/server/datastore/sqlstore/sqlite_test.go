package sqlstore

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	s = sqliteDB{}
)

func TestConnect(t *testing.T) {
	_, err := s.connect("test.db")
	require.NoError(t, err)
	require.FileExists(t, "test.db")

}

func TestAddQuery(t *testing.T) {
	result, err := s.addQueryValues("test.db")
	require.NoError(t, err)
	require.Equal(t, "file:test.db?_foreign_keys=ON&_journal_mode=WAL", result)

}
