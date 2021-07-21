package main

import (
	"os"

	"github.com/ueisele/go-docker-utils/cmd"
)

func main() {
	if err := cmd.Execute(Version()); err != nil {
		os.Exit(1)
	}
}
