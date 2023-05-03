package server

import (
	"github.com/HewlettPackard/galadriel/pkg/server/catalog"
	"net"

	"github.com/sirupsen/logrus"
)

// Config conveys configurations for the Galadriel Server
type Config struct {
	// Address of Galadriel Server
	TCPAddress *net.TCPAddr

	// Address of Galadriel Server to be reached locally
	LocalAddress net.Addr

	// Directory to store runtime data
	DataDir string

	// DB Connection string
	DBConnString string

	Logger logrus.FieldLogger

	ProvidersConfig *catalog.ProvidersConfig
}
