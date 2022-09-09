package harvester

import (
	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
	"github.com/HewlettPackard/galadriel/pkg/harvester/catalog"
	"github.com/sirupsen/logrus"
	"net"
)

// Config conveys configurations for the Harvester.
type Config struct {
	// Address of Harvester
	TCPAddress *net.TCPAddr

	// Address of Harvester to be reached locally
	LocalAddress net.Addr

	// Address of Galadriel server
	ServerAddress string

	// Address of SPIRE Server
	SpireAddress net.Addr

	// Join token to use for attestation
	JoinToken string

	// Directory to store runtime data
	DataDir string

	Log logrus.FieldLogger

	metrics telemetry.MetricServer

	catalog catalog.Catalog
}
