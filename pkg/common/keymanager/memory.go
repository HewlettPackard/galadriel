package keymanager

// Memory is a key manager that keeps keys in memory.
type Memory struct {
	*base
}

func NewMemoryKeyManager(generator Generator) *Memory {
	return &Memory{
		base: newBase(&Config{
			Generator: generator,
		}),
	}
}
