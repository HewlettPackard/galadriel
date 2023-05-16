package server

import (
	"github.com/HewlettPackard/galadriel/pkg/server/catalog"
	"net"

	"github.com/sirupsen/logrus"
)

// Config conveys configurations for the Galadriel Server
type Config struct {
	// TCPAddress represents the TCP address of the Galadriel server.
	TCPAddress *net.TCPAddr

	// LocalAddress represents the Unix Domain Socket (UDS) address of the Server.
	LocalAddress net.Addr

	// DB Connection string
	DBConnString string

	Logger logrus.FieldLogger

	ProvidersConfig *catalog.ProvidersConfig
}
