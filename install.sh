#!/usr/bin/env bash

set -euo pipefail

function usage() {
    echo "Usage: bash install.sh [ -d directory ] [ -v version ]"
    exit 2
}

INSTALL_DIR="${HOME}/bin"
VERSION=""
NAME=lingualeo
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | tr '[:upper:]' '[:lower:]')

case "${OS}" in
linux) OS="linux" ;;
darwin) OS="darwin" ;;
msys* | mingw* | cygwin* | windows_nt) OS="windows" ;;
*)
    echo "Unsupported OS: ${OS}"
    exit 1
    ;;
esac

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
    RELEASE_API_URL="https://api.github.com/repos/trezorg/${NAME}/releases/latest"
else
    RELEASE_API_URL="https://api.github.com/repos/trezorg/${NAME}/releases/tags/${VERSION}"
fi

RELEASE_JSON=$(curl -sSL --fail-with-body "${RELEASE_API_URL}")

if [ -z "${VERSION}" ]; then
    VERSION=$(printf '%s\n' "${RELEASE_JSON}" | awk -F '"' '/tag_name/ { print $4; exit }')
fi

os_pattern="${OS}"
arch_pattern="${ARCH}"

case "${OS}" in
linux) os_pattern="linux|Linux" ;;
darwin) os_pattern="darwin|Darwin" ;;
windows) os_pattern="windows|Windows" ;;
esac

case "${ARCH}" in
amd64) arch_pattern="amd64|x86_64" ;;
arm64) arch_pattern="arm64|aarch64" ;;
esac

DOWNLOAD_URL=$(printf '%s\n' "${RELEASE_JSON}" |
    awk -F '"' '/browser_download_url/ { print $4 }' |
    grep -E "/${NAME}-(${os_pattern})-(${arch_pattern})(\\.exe)?$" |
    head -n 1)

if [ -z "${DOWNLOAD_URL}" ]; then
    echo "Failed to find a matching release asset for ${OS}/${ARCH} in ${VERSION}"
    exit 1
fi

echo "Downloading ${DOWNLOAD_URL} into ${APP_PATH} ..."

if ! curl -sSL --fail-with-body "${DOWNLOAD_URL}" -o "${APP_PATH}"; then
    err=$?
    echo "Failed to download ${DOWNLOAD_URL} into ${APP_PATH}"
    exit ${err}
fi

chmod +x "${APP_PATH}"
"${APP_PATH}" --help || true
