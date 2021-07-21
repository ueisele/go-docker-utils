package cmd

import (
	"fmt"
	"time"
	"golang.org/x/sys/unix"

	"github.com/spf13/cobra"
)

var (
	pathCmd = &cobra.Command{
		Use:   "path",
		Short: "Checks a path on the filesystem for permissions.",
		Long:  "Checks a path on the filesystem for permissions.",
		SilenceUsage: true,
		RunE:  runPathCmd,
	}
	existence   bool
	readable	bool
	writeable   bool
	executable  bool
	timeout		time.Duration
)

func init() {
	pathCmd.Flags().BoolVarP(&existence, "existence", "e", true, "Path must be existence")
	pathCmd.Flags().BoolVarP(&readable, "readable", "r", false, "Path must be readable")
	pathCmd.Flags().BoolVarP(&writeable, "writeable", "w", false, "Path must be writeable")
	pathCmd.Flags().BoolVarP(&executable, "executable", "x", false, "Path must be executable")
	pathCmd.Flags().DurationVarP(&timeout, "timeout", "t", 0, " Time to wait for the URL to be retrievable (default 0s)")
}

func runPathCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires at least one path as argument")
	}
	return checkPathPermissions(cmd.Flags().Args(), flagsToMode(), timeout)
}

func checkPathPermissions(paths []string, mode uint32, timeout time.Duration) error {
	tryUntil := time.Now().Add(timeout)
	for i := 0; i < len(paths); {
		err := unix.Access(paths[i], mode)
		if err == nil {
			i++
		} else if time.Now().Before(tryUntil) {
			time.Sleep(100 * time.Millisecond)
		} else {
			return fmt.Errorf("%v -> %v", paths[i], err)
		}
	}
	return nil
}

func flagsToMode() uint32 {
	var mode uint32
	mode = 0
	if (readable) { mode = mode | unix.R_OK	}
	if (writeable) { mode = mode | unix.W_OK }
	if (executable) { mode = mode | unix.X_OK }
	return mode
}