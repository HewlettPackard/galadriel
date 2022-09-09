package endpoints

import (
	"github.com/HewlettPackard/galadriel/pkg/harvester/catalog"
	"github.com/sirupsen/logrus"
	"net"
)

// Config represents the configuration of the Harvester Endpoints.
type Config struct {
	// TPCAddr is the address to bind the TCP listener to.
	TCPAddress *net.TCPAddr

	// LocalAddr is the local address to bind the listener to.
	LocalAddress net.Addr

	// Plugin catalog
	Catalog catalog.Catalog

	Log logrus.FieldLogger
}
