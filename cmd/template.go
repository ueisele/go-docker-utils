package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ueisele/go-docker-utils/pkg/signals"
	//"github.com/ueisele/go-docker-utils/pkg/template"
)

var (
	templateCmd = &cobra.Command{
		Use:   "template",
		Short: "Uses go template and environment variables to generate configuration files.",
		Long:  "Uses go template and environment variables to generate configuration files.",
		RunE:  runTemplateCmd,
	}
)

func init() {
}

func runTemplateCmd(cmd *cobra.Command, args []string) error {
	stopCh := signals.SetupSignalHandler()

	go func() {
		<-stopCh
		//simulation.Stop()
	}()

	return nil
}
