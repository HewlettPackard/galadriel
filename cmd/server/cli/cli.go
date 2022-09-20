package cli

import (
	"github.com/sirupsen/logrus"
)

type serverCLI struct {
	log logrus.FieldLogger
}

var ServerCLI = &serverCLI{
	log: logrus.New(),
}

func Run() int {
	return ServerCLI.Run()
}

func (c *serverCLI) Run() int {
	return Execute()
}
