#!/usr/bin/env sh

BIN_NAME="${1:-go-docker-utils}"

CGO_ENABLED=0 GOOS=linux go build -o "${BIN_NAME}" -ldflags "-s -w -X main.AppVersionBuild=$(git rev-parse --short=7 HEAD) -X main.AppVersionMetadata=$(date -u +%s)"
#upx "${BIN_NAME}"