package cli

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{}
var cmdExecute = RootCmd.Execute

type serverCLI struct {
	log logrus.FieldLogger
}

func Run() int {
	serverCLI := &serverCLI{
		log: logrus.New(),
	}

	return serverCLI.Run()
}

func (c *serverCLI) Run() int {
	err := cmdExecute()
	if err != nil {
		return 1
	}
	return 0
}
