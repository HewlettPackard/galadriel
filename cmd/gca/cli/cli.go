package cli

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{}
var cmdExecute = RootCmd.Execute

type gcaCLI struct {
	logger logrus.FieldLogger
}

var ServerCLI = &gcaCLI{
	logger: logrus.New(),
}

func Run() int {
	return ServerCLI.Run()
}

func (c *gcaCLI) Run() int {
	return Execute()
}

func Execute() int {
	err := cmdExecute()
	if err != nil {
		return 1
	}
	return 0
}
