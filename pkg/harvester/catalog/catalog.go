package catalog

import (
	"github.com/HewlettPackard/Galadriel/pkg/harvester/server"
	"github.com/HewlettPackard/Galadriel/pkg/harvester/spire"
)

type Catalog struct {
	Spire  spire.SpireServer
	Server server.GaladrielServer
}
