package keymanager

// Memory is a key manager that keeps keys in memory.
type Memory struct {
	*Base
}

func NewMemoryKeyManager(generator Generator) *Memory {
	return &Memory{
		Base: New(&Config{
			Generator: generator,
		}),
	}
}
