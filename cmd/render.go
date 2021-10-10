package cmd

import (
	"os"
	"fmt"
	"github.com/spf13/cobra"

	"github.com/ueisele/go-docker-utils/pkg/engine"
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
	renderCmd.Flags().StringVarP(&output, "out", "o", "", "The output file or directory. If not provided, it is written to stdout.")
	renderCmd.Flags().StringSliceVarP(&values, "values", "v", []string{}, "Values files (glob pattern).")
	renderCmd.Flags().StringSliceVarP(&refs, "refs", "r", []string{}, "Reference templates (glob pattern).")
	renderCmd.Flags().BoolVarP(&strict, "strict", "s", false, "In strict mode, rendering is aborted on missing field. By default its set to zero.")
}

func runRenderCmd(cmd *cobra.Command, args []string) error {
	renderer := engine.NewRenderer().WithConfig(engine.Config{Strict: strict})
	
	var sourceStream engine.Source
	if len(input) > 0 {
		filenames, err := engine.FileGlobsToFileNames(input...)
		if err != nil {
			return fmt.Errorf("could not parse input glob: %v", err) 
		}
		if len(filenames) == 0 {
			return fmt.Errorf("input globs matches no files: %#q", input)
		}
		sourceStream = engine.FileInputSource(filenames...)
	} else {
		sourceStream = engine.ReaderSource("stdin.gotpl", os.Stdin)
	}
	renderer.From(sourceStream)

	if len(values) > 0 {
		filenames, err := engine.FileGlobsToFileNames(values...)
		if err != nil {
			return fmt.Errorf("could not parse values glob: %v", err) 
		}
		contextStream := engine.FileInputSource(filenames...)
		renderer.WithContext(contextStream)
	}

	if len(refs) > 0 {
		filenames, err := engine.FileGlobsToFileNames(refs...)
		if err != nil {
			return fmt.Errorf("could not parse refs glob: %v", err) 
		}
		refsStream := engine.FileInputSource(filenames...)
		renderer.WithReferenceTemplates(refsStream)
	}

	var sinkStream engine.Transform
	if output != "" {
		info, err := os.Stat(output)
		if err == nil && info.Mode().IsDir() {
			sinkStream, err = engine.DirOutputSink(output, ".gotpl", ".tpl")
		} else {
			sinkStream, err = engine.FileOutputSink(output)
		}
		if err != nil {
			return fmt.Errorf("could not create output sink for %s", output) 	
		}
	} else {
		sinkStream = engine.WriterSink(os.Stdout)
	}
	renderer.To(sinkStream)

	return renderer.Render()
}
