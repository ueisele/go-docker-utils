package main

import (
	"fmt"
	"runtime"
)

// app version set by build flag:
//
//	go build -ldflags "-X cmd.AppVersion=$(git describe --tags --dirty --abbrev=7)"
var AppVersion string = "0.0.0"

// app commit set by build flag:
//
//	go build -ldflags "-X main.AppCommit=$(git rev-parse --short=7 HEAD)"
var AppCommit string = "unknown"

func Version() string {
	return fmt.Sprintf("%s (Commit %s) (%s %s/%s).\nCopyright (c) 2023, Uwe Eisele.",
		AppVersion, AppCommit, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
