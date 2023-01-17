package telemetry

import (
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CallCounter struct {
	metrics Metrics
	key     []string
	labels  []Label
	start   time.Time
	done    bool
	mu      sync.Mutex
}

// StartCall starts a "call", which when finished via Done() will emit timing
// and error related metrics.
func StartCall(metrics Metrics, key string, keyn ...string) *CallCounter {
	return &CallCounter{
		metrics: metrics,
		key:     append([]string{key}, keyn...),
		start:   time.Now(),
	}
}

// AddLabel adds a label to be emitted with the call counter.
func (c *CallCounter) AddLabel(name, value string) {
	c.mu.Lock()
	c.labels = append(c.labels, Label{Name: name, Value: value})
	c.mu.Unlock()
}

// Done finishes the "call" and emits metrics.
func (c *CallCounter) Done(errp *error) {
	if c.done {
		return
	}
	c.done = true
	key := c.key

	code := codes.OK
	if errp != nil {
		code = status.Code(*errp)
	}
	c.AddLabel(Status, code.String())

	c.metrics.IncrCounterWithLabels(key, 1, c.labels)
	c.metrics.MeasureSinceWithLabels(append(key, ElapsedTime), c.start, c.labels)
}
