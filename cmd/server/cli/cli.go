package cli

import (
	"github.com/sirupsen/logrus"
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
