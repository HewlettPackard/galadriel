package api

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPServer_Run(t *testing.T) {
	var wg sync.WaitGroup
	s := NewHTTPServer()

	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	var err error
	go func() {
		err = s.Run(ctx)
		wg.Done()
	}()
	cancel()
	wg.Wait()

	assert.NoError(t, err, "unexpected error")
}
