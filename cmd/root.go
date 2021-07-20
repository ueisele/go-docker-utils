package cmd

import (
	"github.com/spf13/cobra"
)
var (
	rootCmd = &cobra.Command{
		Use:	"godub",
		Short:	"godub is a tool which contains a set of utility functions helpful for running docker containers.",
		Long:	"godub is a tool inspired by the Confluent Docker utility belt which contains a set of utility functions helpful for running docker containers.",
	}
)

func Execute(version string) error {
	rootCmd.Version = version
	return rootCmd.Execute()
}

func init()  {
	rootCmd.AddCommand(templateCmd)
}