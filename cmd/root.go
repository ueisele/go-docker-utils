package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:               "godub",
		Short:             "GoDub is a tool which contains a set of utility functions helpful for running containers.",
		Long:              "GoDub is a tool inspired by the Confluent Docker utility belt which contains a set of utility functions helpful for running containers.",
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}
)

func Execute(version string) error {
	rootCmd.Version = version
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(renderCmd)
	rootCmd.AddCommand(ensureCmd)
	rootCmd.AddCommand(pathCmd)
}
