package endpoints

import (
	"net"

	"github.com/sirupsen/logrus"
)

// Config represents the configuration of the Harvester Endpoints.
type Config struct {
	// TPCAddr is the address to bind the TCP listener to.
	TCPAddress *net.TCPAddr

	// LocalAddr is the local address to bind the listener to.
	LocalAddress net.Addr

	Log logrus.FieldLogger
	Logger logrus.FieldLogger
}
