package main

import (
	"fmt"
	"os"

	"github.com/ueisele/go-docker-utils/cmd"
)

func main() {
	if err := cmd.Execute(Version()); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
