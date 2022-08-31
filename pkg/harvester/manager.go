package harvester

import (
	"context"
	"fmt"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/harvester/api"
	"github.com/HewlettPackard/galadriel/pkg/harvester/catalog"
	"github.com/HewlettPackard/galadriel/pkg/harvester/config"
	"github.com/HewlettPackard/galadriel/pkg/harvester/controller"
	"github.com/HewlettPackard/galadriel/pkg/harvester/server"
	"github.com/HewlettPackard/galadriel/pkg/harvester/spire"
)

// Manager is the entity that enables managing the Galadriel Server
type Manager struct {
	catalog    catalog.Catalog
	controller controller.HarvesterController
	api        api.API
	logger     common.Logger
	telemetry  telemetry.MetricServer
}

func NewHarvesterManager() *Manager {
	return &Manager{
		logger: *common.NewLogger(telemetry.Harvester),
	}
}

func (m *Manager) Start(ctx context.Context, config config.HarvesterConfig) {
	type key string

	if m.load(config) != nil {
		panic("unable to load configuration")
	}

	defer m.Stop()

	ctxKey := key(telemetry.PackageName)
	ctx = context.WithValue(ctx, ctxKey, telemetry.Harvester)

	m.run(ctx)
}

func (m *Manager) Stop() {
	// unload and cleanup stuff
}

func (m *Manager) load(config config.HarvesterConfig) error {
	cat := catalog.Catalog{
		Spire:  spire.NewLocalSpireServer(config.HarvesterConfigSection.SpireSocketPath),
		Server: server.NewRemoteGaladrielServer(config.HarvesterConfigSection.ServerAddress),
	}

	controller := controller.NewLocalHarvesterController(cat)
	api := api.NewHTTPApi(controller)

	m.catalog = cat
	m.controller = controller
	m.api = api

	telemetry := telemetry.NewLocalMetricServer()
	m.telemetry = telemetry

	return nil
}

func (m *Manager) run(ctx context.Context) {
	m.logger.Info("Starting harvester manager")

	var wg sync.WaitGroup

	_, cancel := context.WithCancel(ctx)
	ctx, _ = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	defer func() {
		wg.Wait()
		cancel()
	}()

	plugins := []common.RunnablePlugin{
		m.controller,
		m.api,
		m.telemetry,
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
