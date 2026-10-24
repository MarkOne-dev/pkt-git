#!/usr/bin/env bash
# install.sh — install the pkt-git Go binary on this machine.
# Repositories are then wired with a single `pkt init`.
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd -- "$SCRIPT_DIR/.." && pwd)"
BIN_DIR="${PKTG_BIN_DIR:-${HOME}/.local/bin}"
TARGET="${BIN_DIR}/pkt"

TMP_BUILD=""
TMP_TARGET=""
cleanup() {
    [[ -n "${TMP_BUILD}" ]] && rm -rf "${TMP_BUILD}"
    [[ -n "${TMP_TARGET}" ]] && rm -f "${TMP_TARGET}"
}
trap cleanup EXIT

find_prebuilt() {
    for candidate in "${ROOT_DIR}/bin/pkt" "${ROOT_DIR}/pkt"; do
        if [[ -x "${candidate}" ]]; then
            printf '%s\n' "${candidate}"
            return 0
        fi
    done
    return 1
}

SOURCE="${PKTG_SOURCE:-}"
if [[ -z "${SOURCE}" ]]; then
    if ! SOURCE="$(find_prebuilt)"; then
        if ! command -v go >/dev/null 2>&1; then
            echo "pkt-git: no prebuilt binary found and Go is not installed." >&2
            echo "Build with 'make build' or install a release archive." >&2
            exit 1
        fi
        TMP_BUILD="$(mktemp -d "${TMPDIR:-/tmp}/pkt-git-build.XXXXXX")"
        (cd "${ROOT_DIR}" && CGO_ENABLED=0 go build -o "${TMP_BUILD}/pkt" ./cmd/pkt)
        SOURCE="${TMP_BUILD}/pkt"
    fi
fi

if [[ ! -f "${SOURCE}" ]]; then
    echo "pkt-git: binary not found: ${SOURCE}" >&2
    exit 1
fi

mkdir -p "${BIN_DIR}"
TMP_TARGET="$(mktemp "${BIN_DIR}/.pkt.XXXXXX")"
install -m 0755 "${SOURCE}" "${TMP_TARGET}"
mv -f "${TMP_TARGET}" "${TARGET}"
TMP_TARGET=""

"${TARGET}" version

if [[ "${PKTG_SKIP_SKILLS:-0}" != "1" ]]; then
    "${TARGET}" skills
fi

case ":${PATH}:" in
    *":${BIN_DIR}:"*) ;;
    *)
        echo "note: ${BIN_DIR} is not on PATH." >&2
        echo "      Add 'export PATH=\"\$PATH:${BIN_DIR}\"' to your shell profile." >&2
        ;;
esac

echo "pkt-git installed at ${TARGET}."
echo "In any Git repository run:"
echo "    pkt init          # wire hooks and configuration"
echo "    pkt --help        # full command list"
