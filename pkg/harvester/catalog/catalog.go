package catalog

import (
	"github.com/HewlettPackard/galadriel/pkg/harvester/server"
	"github.com/HewlettPackard/galadriel/pkg/harvester/spire"
)

type Catalog struct {
	Spire  spire.SpireServer
	Server server.GaladrielServer
}
