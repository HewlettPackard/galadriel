package endpoints

import (
	"github.com/HewlettPackard/galadriel/pkg/common/ca"
	"net"

	"github.com/sirupsen/logrus"
)

// Config represents the configuration of the Galadriel Server Endpoints
type Config struct {
	// CA is used for signing X.509 certificates
	CA *ca.CA

	// TPCAddr is the address to bind the TCP listener to.
	TCPAddress *net.TCPAddr

	// LocalAddress is the local address to bind the listener to.
	LocalAddress net.Addr

	// Postgres connection string
	DatastoreConnString string

	Logger logrus.FieldLogger
}
