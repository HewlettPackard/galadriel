package harvester

import (
	"net"

	"github.com/sirupsen/logrus"
)

const (
	postBundlePath  = "/bundle"
	syncBundlesPath = "/bundle/sync"
	connectPath     = "/onboard"
)

type Config struct {
	TCPAddress *net.TCPAddr
	Logger     logrus.FieldLogger
}
