package endpoints

import (
	"net"

	"github.com/sirupsen/logrus"
)

// Config represents the configuration of the Harvester Endpoints.
type Config struct {
	// TPCAddr is the address to bind the TCP listener to.
	TCPAddress *net.TCPAddr

	// localAddr is the local address to bind the listener to.
	LocalAddress net.Addr

	Logger logrus.FieldLogger
}
