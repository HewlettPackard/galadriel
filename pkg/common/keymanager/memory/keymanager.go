package memory

import (
	"github.com/HewlettPackard/galadriel/pkg/common/keymanager"
)

// KeyManager is a key manager that keeps keys in memory.
type KeyManager struct {
	*keymanager.Base
}

// New creates a new memory key manager.
func New(generator keymanager.Generator) *KeyManager {
	return &KeyManager{
		Base: keymanager.New(&keymanager.Config{
			Generator: generator,
		}),
	}
}
