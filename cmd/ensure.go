package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	ensureCmd = &cobra.Command{
		Use:   "ensure",
		Short: "Uses go template and environment variables to generate configuration files.",
		Long:  "Uses go template and environment variables to generate configuration files.",
		SilenceUsage: true,
		RunE:  runEnsureCmd,
	}
	atLeastOne bool
)

func init() {
	ensureCmd.Flags().BoolVarP(&atLeastOne, "at-least-one", "a", false, "By the default it is ensured that all environment variables are defined. If this flag is set, it is enough if at least one is defined.")
}

func runEnsureCmd(cmd *cobra.Command, args []string) error {
	var err error
	if atLeastOne {
		err = checkAtLeastOnePresent(cmd.Flags().Args())
	} else {
		err = checkAllPresent(cmd.Flags().Args())
	}
	return err
}

func checkAllPresent(envs []string) error {
	missingEnvs := make([]string, 0)
	for _, env := range envs {
		if len(os.Getenv(env)) == 0 {
			missingEnvs = append(missingEnvs, env)
		}
	}
	if len(missingEnvs) > 0 {
		return fmt.Errorf("environment variables are missing: %v", missingEnvs)
	}
	return nil
}

func checkAtLeastOnePresent(envs []string) error {
	for _, env := range envs {
		if len(os.Getenv(env)) > 0 {
			return nil
		}
	}
	return fmt.Errorf("none of the specified environment variables is present: %v", envs)
}