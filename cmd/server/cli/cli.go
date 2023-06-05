package cli

import (
	"github.com/sirupsen/logrus"
)

const (
	defaultSocketPath = "/tmp/galadriel-server/api.sock"
	defaultConfigPath = "conf/server/server.conf"
)

type serverCLI struct {
	logger logrus.FieldLogger
}

var ServerCLI = &serverCLI{
	logger: logrus.New(),
}

func Run() int {
	return ServerCLI.Run()
}

func (c *serverCLI) Run() int {
	return Execute()
}
