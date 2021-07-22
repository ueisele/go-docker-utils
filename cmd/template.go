package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ueisele/go-docker-utils/pkg/template"
)

var (
	templateCmd = &cobra.Command{
		Use:   "template",
		Short: "Uses Go template and environment variables to generate configuration files.",
		Long:  "Uses Go template and environment variables to generate configuration files.",
		SilenceUsage: true,
		RunE:  runTemplateCmd,
	}
	in 			string
	out 		string
	missingkey  string
)

func init() {
	templateCmd.Flags().StringVarP(&in, "in", "i", "", "The template file. If not provided, it is read from stdin.")
	templateCmd.Flags().StringVarP(&out, "out", "o", "", "The output file. If not provided, it is written from stdout.")
	templateCmd.Flags().StringVarP(&missingkey, "missingkey", "m", "default", "Strategy for dealing with missing keys: [default|zero|error]")
}

func runTemplateCmd(cmd *cobra.Command, args []string) error {
	var tplFile *os.File
	var err error
	if len(in) > 0 {
		tplFile, err = os.Open(in)
		if err != nil {
			return fmt.Errorf("could not open template file %s: %v", in, err)
		}
		defer tplFile.Close()
	} else {
		tplFile = os.Stdin
	}

	var outFile *os.File
	if len(out) > 0 {
		outFile, err = os.Create(out)
		if err != nil {
			return fmt.Errorf("could not create output file %s: %v", out, err)
		}
		defer outFile.Close()
	} else {
		outFile = os.Stdout
	}

	dubtemplate := template.NewDubTemplateWithDefaults("missingkey=" + missingkey)
	return dubtemplate.TemplateOsFileToOsFile(tplFile, outFile)
}
