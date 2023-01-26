#!/usr/bin/env sh

set -eou pipefail

function usage() {
    echo "Usage: $0 [ -d directory ]"
    exit 2
}

INSTALL_DIR="${HOME}/bin"
NAME=lingualeo
OS=$(uname -o | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | tr '[:upper:]' '[:lower:]')

if which go &>/dev/null; then
    gobin=$(go env GOBIN)
    gopath="$(go env GOPATH)"
    gopathbin="${gopath}/bin"
    if [ -n "${gobin}" ] && [ -d "${gobin}" ]; then
        INSTALL_DIR="${gobin}"
    elif [ -n "${gopath}" ] && [ -d ${gopathbin} ]; then
        INSTALL_DIR="${gopathbin}"
    fi
fi

while getopts hd: flag; do
    case "${flag}" in
    d) INSTALL_DIR=${OPTARG} ;;
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

echo "Installalling into ${APP_PATH}..."

TAG_NAME=$(curl -s https://api.github.com/repos/trezorg/${NAME}/releases/latest | awk -F ':' '/tag_name/ { gsub("[\", ]", "", $2); print $2 }')
DOWNLOAD_URL="https://github.com/trezorg/${NAME}/releases/download/${TAG_NAME}/${NAME}-${OS}-${ARCH}"

if ! curl --fail-with-body -sL "${DOWNLOAD_URL}" -o "${APP_PATH}"; then
    err=$?
    echo "Failed to download ${DOWNLOAD_URL} into ${APP_PATH}"
    exit ${err}
fi

chmod +x "${APP_PATH}"
"${NAME}" --help
echo $?
