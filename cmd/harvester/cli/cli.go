package cli

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/HewlettPackard/galadriel/pkg/common/telemetry"
)

var (
	RootCmd      = NewRootCmd()
	cmdExecute   = RootCmd.Execute
	HarvesterCmd = &HarvesterCLI{
		logger: logrus.WithField(telemetry.SubsystemName, telemetry.Harvester),
	}
)

type HarvesterCLI struct {
	logger logrus.FieldLogger
}

func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "harvester",
		Long: "This is the Galadriel Harvester CLI",
	}
}

func Run() int {
	err := HarvesterCmd.Run()
	if err != nil {
		return 1
	}

	return 0
}

func (hc *HarvesterCLI) Run() error {
	return hc.Execute()
}

func (*HarvesterCLI) Execute() error {
	err := cmdExecute()
	if err != nil {
		return err
	}

	return nil
}
