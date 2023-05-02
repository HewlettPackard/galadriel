package harvester

import (
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

// Config conveys configurations for the Harvester.
type Config struct {
	// Address of Harvester
	TCPAddress *net.TCPAddr

	// Address of Harvester to be reached locally
	LocalAddress net.Addr

	// Address of Galadriel server
	ServerAddress *net.TCPAddr

	// Address of SPIRE Server
	SpireAddress net.Addr

	// Access token for connecting to Galadriel Server
	JoinToken string

	// How often to check for bundle rotation
	BundleUpdatesInterval time.Duration

	// Directory to store runtime data
	DataDir string

	TrustBundlePath string

	Logger logrus.FieldLogger
}
