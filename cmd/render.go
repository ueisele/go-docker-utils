package cmd

import (
	"github.com/spf13/cobra"

	//"github.com/ueisele/go-docker-utils/pkg/engine"
)

var (
	renderCmd = &cobra.Command{
		Use:   "render",
		Short: "Uses Go template and environment variables to generate configuration files.",
		Long:  "Uses Go template and environment variables to generate configuration files.",
		SilenceUsage: true,
		RunE:  runRenderCmd,
	}
	input 		[]string
	output 		string
	values		[]string
	refs  		[]string
	strict 		bool
)

func init() {
	renderCmd.Flags().StringSliceVarP(&input, "in", "i", []string{}, "The template files (glob pattern). If not provided, it is read from stdin.")
	renderCmd.Flags().StringVarP(&output, "out", "o", "", "The output file or directory. If not provided, it is written from stdout.")
	renderCmd.Flags().StringSliceVarP(&values, "values", "v", []string{}, "Values files (glob pattern).")
	renderCmd.Flags().StringSliceVarP(&refs, "refs", "r", []string{}, "Reference templates (glob pattern).")
	renderCmd.Flags().BoolVarP(&strict, "strict", "s", false, "In strict mode, rendering is aborted on missing field. By default its set to zero.")
}

func runRenderCmd(cmd *cobra.Command, args []string) error {
	return nil
}
