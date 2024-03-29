package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"

	"github.com/ueisele/go-docker-utils/pkg/template"
)

var (
	renderCmd = &cobra.Command{
		Use:          "template",
		Short:        "Uses Go template and environment variables to generate configuration files.",
		Long:         "Uses Go template and environment variables to generate configuration files.",
		SilenceUsage: true,
		RunE:         runRenderCmd,
	}
	input  []string
	output string
	refs   []string
	values []string
	files  []string
	strict bool
)

func init() {
	renderCmd.Flags().StringSliceVarP(&input, "in", "i", []string{}, "The template files (glob pattern). If not provided, it is read from stdin.")
	renderCmd.Flags().StringVarP(&output, "out", "o", "", "The output file or directory. If not provided, it is written to stdout.")
	renderCmd.Flags().StringSliceVarP(&refs, "refs", "r", []string{}, "Reference templates (glob pattern).")
	renderCmd.Flags().StringSliceVarP(&values, "values", "v", []string{}, "Values files (glob pattern). Can be used with '.Values.' prefix.")
	renderCmd.Flags().StringSliceVarP(&files, "files", "f", []string{},
		"Available files inside templates (directories). It should be noted that all files are immediately loaded into memory. Can be used with '.Files.' prefix.")
	renderCmd.Flags().BoolVarP(&strict, "strict", "s", false, "In strict mode, rendering is aborted on missing field.")
}

func runRenderCmd(cmd *cobra.Command, args []string) error {
	renderer := template.NewRenderer().WithConfig(template.Config{Strict: strict})

	var sourceStream template.Source
	if len(input) > 0 {
		filenames, err := template.FileGlobsToFileNames(input...)
		if err != nil {
			return fmt.Errorf("could not parse input glob: %v", err)
		}
		if len(filenames) == 0 {
			return fmt.Errorf("input globs matches no files: %#q", input)
		}
		sourceStream = template.FileInputSource(filenames...)
	} else {
		sourceStream = template.ReaderSource("stdin.gotpl", os.Stdin)
	}
	renderer.From(sourceStream)

	if len(refs) > 0 {
		filenames, err := template.FileGlobsToFileNames(refs...)
		if err != nil {
			return fmt.Errorf("could not parse refs glob: %v", err)
		}
		refsStream := template.FileInputSource(filenames...)
		renderer.WithReferenceTemplates(refsStream)
	}

	if len(values) > 0 {
		filenames, err := template.FileGlobsToFileNames(values...)
		if err != nil {
			return fmt.Errorf("could not parse values glob: %v", err)
		}
		valuesStream := template.FileInputSource(filenames...)
		renderer.WithValues(valuesStream)
	}

	for _, filesDir := range files {
		stat, err := os.Stat(filesDir)
		if err != nil {
			return fmt.Errorf("'%v' cannot be used for files, because stat failed because of %v", filesDir, err.Error())
		}
		if !stat.IsDir() {
			return fmt.Errorf("'%v' cannot be used for files, because only directories are allowed", filesDir)
		}
		filesStream := template.DirInputSource(os.DirFS(filesDir))
		renderer.WithFiles(filesStream)
	}

	var sinkStream template.Transform
	if output != "" {
		info, err := os.Stat(output)
		if err == nil && info.Mode().IsDir() {
			sinkStream, err = template.DirOutputSink(output, ".gotpl", ".tpl")
		} else {
			sinkStream, err = template.FileOutputSink(output)
		}
		if err != nil {
			return fmt.Errorf("could not create output sink for %s", output)
		}
	} else {
		sinkStream = template.WriterSink(os.Stdout)
	}
	renderer.To(sinkStream)

	return renderer.Render()
}
