#!/usr/bin/env bash

set -euo pipefail

function usage() {
    echo "Usage: bash install.sh [ -d directory ] [ -v version ]"
    exit 2
}

INSTALL_DIR="${HOME}/bin"
VERSION=""
NAME=lingualeo
OS=$(uname -o | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | tr '[:upper:]' '[:lower:]')
OS="${OS##*/}"

case "${ARCH}" in
x86_64) ARCH="amd64" ;;
aarch64 | arm64 | arm) ARCH="arm64" ;;
armv*) ARCH="arm" ;;
esac

if which go &>/dev/null; then
    gobin=$(go env GOBIN)
    gopath="$(go env GOPATH)"
    gopathbin="${gopath}/bin"
    if [ -n "${gobin}" ] && [ -d "${gobin}" ]; then
        INSTALL_DIR="${gobin}"
    elif [ -n "${gopath}" ] && [ -d "${gopathbin}" ]; then
        INSTALL_DIR="${gopathbin}"
    fi
fi

while getopts hv:d: flag; do
    case "${flag}" in
    d) INSTALL_DIR=${OPTARG} ;;
    v) VERSION=${OPTARG} ;;
    h) usage ;;
    *) usage ;;
    esac
done

if [ ! -d "${INSTALL_DIR}" ]; then
    echo "Directory ${INSTALL_DIR} does not exist"
    exit 1
fi
if [ ! -w "${INSTALL_DIR}" ]; then
    echo "Directory ${INSTALL_DIR} is not writable"
    exit 1
fi
APP_PATH="${INSTALL_DIR}/${NAME}"

# Cleanup on failure
trap 'rm -f "${APP_PATH}"' ERR

echo "Installing into ${APP_PATH}..."

if [ -z "${VERSION}" ]; then
    VERSION=$(
        curl -sSL --fail-with-body https://api.github.com/repos/trezorg/${NAME}/releases/latest |
            awk -F '"' '/tag_name/ { print $4 }'
    )
fi
DOWNLOAD_URL="https://github.com/trezorg/${NAME}/releases/download/${VERSION}/${NAME}-${OS}-${ARCH}"
echo "Downloading ${DOWNLOAD_URL}..."

if ! curl -sSL --fail-with-body "${DOWNLOAD_URL}" -o "${APP_PATH}"; then
    err=$?
    echo "Failed to download ${DOWNLOAD_URL} into ${APP_PATH}"
    exit ${err}
fi

chmod +x "${APP_PATH}"
"${APP_PATH}" --help
