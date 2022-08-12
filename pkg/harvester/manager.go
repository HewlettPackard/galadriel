package harvester

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/HewlettPackard/galadriel/pkg/common"
	"github.com/HewlettPackard/galadriel/pkg/harvester/api"
	"github.com/HewlettPackard/galadriel/pkg/harvester/catalog"
	"github.com/HewlettPackard/galadriel/pkg/harvester/config"
	"github.com/HewlettPackard/galadriel/pkg/harvester/controller"
	"github.com/HewlettPackard/galadriel/pkg/harvester/server"
	"github.com/HewlettPackard/galadriel/pkg/harvester/spire"
)

type HarvesterManager struct {
	catalog    catalog.Catalog
	controller controller.HarvesterController
	api        api.API
	logger     common.Logger
}

func NewHarvesterManager() *HarvesterManager {
	return &HarvesterManager{
		logger: *common.NewLogger("harvester_manager"),
	}
}

func (m *HarvesterManager) Start(ctx context.Context, config config.HarvesterConfig) {
	if m.load(config) != nil {
		panic("unable to load configuration")
	}

	defer m.Stop()
	m.run(ctx)
}

func (m *HarvesterManager) Stop() {
	// unload and cleanup stuff
}

func (m *HarvesterManager) load(config config.HarvesterConfig) error {
	cat := catalog.Catalog{
		Spire:  spire.NewLocalSpireServer(config.HarvesterConfigSection.SpireSocketPath),
		Server: server.NewRemoteGaladrielServer(config.HarvesterConfigSection.ServerAddress),
	}
	controller := controller.NewLocalHarvesterController(cat)
	api := api.NewHTTPApi(controller)

	m.catalog = cat
	m.controller = controller
	m.api = api

	return nil
}

func (m *HarvesterManager) run(ctx context.Context) {
	// TODO: figure out how to trap signals
	m.logger.Info("Starting harvester manager")

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		wg.Wait()
		cancel()
	}()

	plugins := []common.RunnablePlugin{
		m.controller,
		m.api,
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
