package harvester

import (
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

// Config conveys configurations for the Harvester.
type Config struct {
	// LocalAddress represents the Unix Domain Socket (UDS) address of the Harvester.
	LocalAddress net.Addr

	// TCP Address of Galadriel server
	ServerAddress *net.TCPAddr

	// Unix Domain Socket (UDS) Address of SPIRE Server
	LocalSpireAddress net.Addr

	// Access token for connecting to Galadriel Server
	JoinToken string

	// How often to check for bundle rotation
	BundleUpdatesInterval time.Duration

	// Path to the trust bundle for the Galadriel Server
	ServerTrustBundlePath string

	Logger logrus.FieldLogger
}
