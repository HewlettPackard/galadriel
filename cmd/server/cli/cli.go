package cli

import (
	"github.com/HewlettPackard/galadriel/pkg/common"
)

type serverCLI struct {
	logger *common.Logger
}

var ServerCLI = &serverCLI{
	logger: common.NewLogger("server"),
}

func Run() int {
	return ServerCLI.Run()
}

func (c *serverCLI) Run() int {
	c.logger.Info("Starting the Server CLI")
	return Execute()
}
