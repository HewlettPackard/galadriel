package server

import (
	"context"
	"fmt"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/server/api"
	"github.com/HewlettPackard/galadriel/pkg/server/config"
)

// Manager is the entity that enables managing the Galadriel Server
type Manager struct {
	api       api.HTTPServer
	config    config.Server
	logger    common.Logger
	telemetry telemetry.MetricServer
}

// NewManager returns a new Galadriel Server Manager
func NewManager() *Manager {
	return &Manager{
		logger: *common.NewLogger(telemetry.Server),
	}
}

// Run runs the Galadriel server.
func Run(configPath string) error {
	m := NewManager()
	m.logger.Info("Starting the Galadriel Server")

	if err := m.load(configPath); err != nil {
		m.logger.Error("Error configuring server:", err)
		return err
	}

	m.Start()

	return nil
}

func (m *Manager) Start() {
	type key string
	defer m.Stop()

	ctxKey := key(telemetry.PackageName)
	ctx := context.WithValue(context.Background(), ctxKey, telemetry.Harvester)

	m.run(ctx)
}

func (m *Manager) Stop() {
	return
}

func (m *Manager) load(configPath string) error {
	c, err := config.LoadFromDisk(configPath)
	if err != nil {
		return err
	}
	m.config = *c
	m.api = api.NewHTTPServer()
	return nil
}

func (m *Manager) run(ctx context.Context) {
	m.logger.Info("Starting Server manager")

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	ctx, _ = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		wg.Wait()
		cancel()
	}()

	plugins := []common.RunnablePlugin{
		api.NewHTTPServer(),
	}
	wg.Add(len(plugins))

	errch := make(chan error, len(plugins))
	for _, plugin := range plugins {
		plugin := plugin
		go func() {
			errch <- runTask(ctx, &wg, plugin.Run)
		}()
	}
}

func runTask(ctx context.Context, wg *sync.WaitGroup, task func(context.Context) error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v\n%s", r, string(debug.Stack()))
		}
		wg.Done()
	}()

	return task(ctx)
}
