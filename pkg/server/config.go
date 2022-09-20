package server

import (
	"github.com/sirupsen/logrus"
	"net"
)

// Config conveys configurations for the Galadriel Server
type Config struct {
	// Address of Galadriel Server
	TCPAddress *net.TCPAddr

	// Address of Galadriel Server to be reached locally
	LocalAddress net.Addr

	// Directory to store runtime data
	DataDir string

	Log logrus.FieldLogger
}
