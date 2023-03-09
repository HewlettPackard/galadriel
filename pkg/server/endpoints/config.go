package endpoints

import (
	"net"

	"github.com/sirupsen/logrus"
)

// Config represents the configuration of the Galadriel Server Endpoints
type Config struct {
	// TPCAddr is the address to bind the TCP listener to.
	TCPAddress *net.TCPAddr

	// CertPath for server's TCP endpoint certificate.
	CertPath string

	// CertKeyPath for server's TCP endpoint certificate key.
	CertKeyPath string

	// LocalAddress is the local address to bind the listener to.
	LocalAddress net.Addr

	// Postgres connection string
	DatastoreConnString string

	Logger logrus.FieldLogger
}
