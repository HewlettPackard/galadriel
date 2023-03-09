package server

import (
	"net"

	"github.com/sirupsen/logrus"
)

// Config conveys configurations for the Galadriel Server
type Config struct {
	// TCPAddress of Galadriel Server
	TCPAddress *net.TCPAddr

	// LocalAddress of Galadriel Server to be reached locally
	LocalAddress net.Addr

	// CertPath for server's certificate. Used for harvester TLS connection.
	CertPath string

	// CertKeyPath for server's certificate key. Used for harvester TLS connection
	CertKeyPath string

	// DataDir is the directory path to store runtime data
	DataDir string

	// DBConnString DB Connection string
	DBConnString string

	Logger logrus.FieldLogger
}
