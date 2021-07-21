# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.188.0/containers/go/.devcontainer/base.Dockerfile
ARG VARIANT="1"
FROM mcr.microsoft.com/vscode/devcontainers/go:0-${VARIANT}

RUN apt-get install -y upx \
    && apt-get clean