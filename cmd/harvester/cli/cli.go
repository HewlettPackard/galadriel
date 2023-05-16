package cli

import (
	"github.com/sirupsen/logrus"
)

type harvesterCLI struct {
	logger logrus.FieldLogger
}

var ServerCLI = &harvesterCLI{
	logger: logrus.New(),
}

func Run() int {
	return ServerCLI.Run()
}

func (c *harvesterCLI) Run() int {
	return Execute()
}
