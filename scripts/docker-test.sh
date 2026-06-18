#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${ROOT_DIR}/dist"
APP="config-cli"

IMAGE="${IMAGE:-debian:12}"
DOCKER_TTY="${DOCKER_TTY:--it}"

pick_bin_for_uname() {
  local u="$1"
  case "$u" in
    aarch64|arm64) echo "${DIST_DIR}/${APP}-linux-arm64" ;;
    x86_64|amd64) echo "${DIST_DIR}/${APP}-linux-amd64" ;;
    *) echo "" ;;
  esac
}

if [[ -n "${TARGET_ARCH:-}" ]]; then
  case "${TARGET_ARCH}" in
    arm64|aarch64) BIN_PATH="${DIST_DIR}/${APP}-linux-arm64" ;;
    amd64|x86_64) BIN_PATH="${DIST_DIR}/${APP}-linux-amd64" ;;
    *)
      echo "Invalid TARGET_ARCH=${TARGET_ARCH} (use arm64 or amd64)" >&2
      exit 2
      ;;
  esac
else
  U="$(docker run --rm "${IMAGE}" uname -m)"
  BIN_PATH="$(pick_bin_for_uname "${U}")"
  if [[ -z "${BIN_PATH}" ]]; then
    echo "Unsupported container arch from uname -m: ${U}" >&2
    exit 2
  fi
  echo "Detected container uname -m: ${U}"
  echo "Selected binary: ${BIN_PATH}"
fi

if [[ ! -f "${BIN_PATH}" ]]; then
  echo "Binary not found: ${BIN_PATH}" >&2
  echo "Run: (cd ${ROOT_DIR} && make dist)" >&2
  exit 1
fi

if [[ ! -x "${BIN_PATH}" ]]; then
  chmod +x "${BIN_PATH}" || true
fi

# Mount the selected Linux binary to the *container filesystem root* as /config-cli
exec docker run --rm ${DOCKER_TTY} \
  -v "${ROOT_DIR}:/workspace" \
  -v "${BIN_PATH}:/config-cli" \
  -w / \
  "${IMAGE}" \
  /config-cli
