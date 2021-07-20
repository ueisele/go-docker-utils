
package main

import (
	"fmt"
	"runtime"
)

const (
	AppName         = "godub"
	AppVersionMajor = 0
	AppVersionMinor = 1
	AppVersionPatch = 0
	AppVersionBuild = ""
)

// version build metadata set by build flag:
//     go build -ldflags "-X cmd.AppVersionMetadata=$(date -u +%s)"
var AppVersionMetadata string

func Version() string {
	// major.minor.patch[-prerelease+buildmetadata]
	// optional version suffix format is "-(pre-release-version)+(build-metadata)"
	suffix := ""

	if AppVersionBuild != "" {
		suffix += "-" + AppVersionBuild
	}

	if AppVersionMetadata != "" {
		suffix += "-" + AppVersionMetadata
	}

	return fmt.Sprintf("%s %d.%d.%d%s (Go runtime %s).\nCopyright (c) 2021, Uwe Eisele.",
		AppName, AppVersionMajor, AppVersionMinor, AppVersionPatch, suffix, runtime.Version())
}