#!/usr/bin/env sh

BIN_NAME="${1:-godub}"

CGO_ENABLED=0 go build -o "${BIN_NAME}" -ldflags "-s -w -X main.AppVersion=$(git describe --tags --dirty --abbrev=7) -X main.AppCommit=$(git rev-parse --short=7 HEAD)"
upx -6 "${BIN_NAME}"
