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
    [[ -n "${TMP_BUILD}" ]] && rm -rf "${TMP_BUILD}" || true
    [[ -n "${TMP_TARGET}" ]] && rm -f "${TMP_TARGET}" || true
}
trap cleanup EXIT

install_git_if_missing() {
    if command -v git >/dev/null 2>&1; then
        return 0
    fi

    echo "pkt-git: Git is not installed. Attempting automatic installation..."
    local SUDO=""
    if [[ "${EUID:-$(id -u)}" -ne 0 ]]; then
        if command -v sudo >/dev/null 2>&1; then
            SUDO="sudo"
        fi
    fi

    if command -v apt-get >/dev/null 2>&1; then
        ${SUDO} apt-get update && ${SUDO} apt-get install -y git
    elif command -v dnf >/dev/null 2>&1; then
        ${SUDO} dnf install -y git
    elif command -v pacman >/dev/null 2>&1; then
        ${SUDO} pacman -Sy --noconfirm git
    elif command -v zypper >/dev/null 2>&1; then
        ${SUDO} zypper install -y git
    elif command -v brew >/dev/null 2>&1; then
        brew install git
    else
        echo "warning: could not automatically install Git. Please install Git manually." >&2
        return 0
    fi
}

configure_git_identity_if_needed() {
    if ! command -v git >/dev/null 2>&1; then
        return 0
    fi
    local user_name user_email
    user_name="$(git config user.name 2>/dev/null || true)"
    user_email="$(git config user.email 2>/dev/null || true)"

    if [[ -z "${user_name}" || -z "${user_email}" ]]; then
        if [[ -t 0 && "${CI:-0}" != "1" && "${PKTG_NONINTERACTIVE:-0}" != "1" ]]; then
            echo ""
            echo "Git identity (user.name / user.email) is not configured."
            echo "pkt-git requires a configured Git identity to prevent commit errors."
            if [[ -z "${user_name}" ]]; then
                read -r -p "Enter your full name for Git commits: " user_name
                if [[ -n "${user_name}" ]]; then
                    git config --global user.name "${user_name}"
                fi
            fi
            if [[ -z "${user_email}" ]]; then
                read -r -p "Enter your email for Git commits: " user_email
                if [[ -n "${user_email}" ]]; then
                    git config --global user.email "${user_email}"
                fi
            fi
        fi
    fi
}

configure_path_if_needed() {
    case ":${PATH}:" in
        *":${BIN_DIR}:"*) return 0 ;;
    esac

    echo "note: ${BIN_DIR} is not currently on PATH."
    local export_line="export PATH=\"\${PATH}:${BIN_DIR}\""
    local profile_updated=0

    if [[ -f "${HOME}/.bashrc" ]] && ! grep -Fq "${BIN_DIR}" "${HOME}/.bashrc"; then
        printf '\n# Added by pkt-git installer\n%s\n' "${export_line}" >> "${HOME}/.bashrc"
        echo "  Added ${BIN_DIR} to ~/.bashrc"
        profile_updated=1
    fi
    if [[ -f "${HOME}/.zshrc" ]] && ! grep -Fq "${BIN_DIR}" "${HOME}/.zshrc"; then
        printf '\n# Added by pkt-git installer\n%s\n' "${export_line}" >> "${HOME}/.zshrc"
        echo "  Added ${BIN_DIR} to ~/.zshrc"
        profile_updated=1
    fi
    if [[ -f "${HOME}/.bash_profile" ]] && ! grep -Fq "${BIN_DIR}" "${HOME}/.bash_profile"; then
        printf '\n# Added by pkt-git installer\n%s\n' "${export_line}" >> "${HOME}/.bash_profile"
        echo "  Added ${BIN_DIR} to ~/.bash_profile"
        profile_updated=1
    fi
    if [[ -f "${HOME}/.profile" ]] && ! grep -Fq "${BIN_DIR}" "${HOME}/.profile"; then
        printf '\n# Added by pkt-git installer\n%s\n' "${export_line}" >> "${HOME}/.profile"
        echo "  Added ${BIN_DIR} to ~/.profile"
        profile_updated=1
    fi

    if [[ "${profile_updated}" -eq 0 ]]; then
        echo "      Add 'export PATH=\"\$PATH:${BIN_DIR}\"' to your shell profile." >&2
    fi
}

find_prebuilt() {
    for candidate in "${ROOT_DIR}/bin/pkt" "${ROOT_DIR}/pkt"; do
        if [[ -x "${candidate}" ]]; then
            printf '%s\n' "${candidate}"
            return 0
        fi
    done
    return 1
}

install_nerd_font_if_needed() {
    if [[ "${PKTG_SKIP_FONTS:-0}" == "1" ]]; then
        return 0
    fi

    # Check if a Nerd Font is already present
    if command -v fc-list >/dev/null 2>&1; then
        if fc-list : family 2>/dev/null | grep -qi "Nerd Font"; then
            return 0
        fi
    fi

    local OS
    OS="$(uname -s 2>/dev/null || echo "unknown")"

    if [[ "${OS}" == "Darwin" ]]; then
        if command -v brew >/dev/null 2>&1; then
            echo "pkt-git: Installing JetBrainsMono Nerd Font via Homebrew..."
            brew install --cask font-jetbrains-mono-nerd-font >/dev/null 2>&1 || true
            return 0
        fi
        local mac_font_dir="${HOME}/Library/Fonts"
        if ls "${mac_font_dir}"/*Nerd* >/dev/null 2>&1; then
            return 0
        fi
        echo "pkt-git: Downloading and installing JetBrainsMono Nerd Font..."
        mkdir -p "${mac_font_dir}"
        local tmp_font_dir
        tmp_font_dir="$(mktemp -d "${TMPDIR:-/tmp}/pkt-font.XXXXXX")"
        local font_url="https://github.com/ryanoasis/nerd-fonts/releases/latest/download/JetBrainsMono.tar.xz"
        if curl -fsSL "${font_url}" -o "${tmp_font_dir}/font.tar.xz" 2>/dev/null; then
            tar -xf "${tmp_font_dir}/font.tar.xz" -C "${tmp_font_dir}" 2>/dev/null || true
            find "${tmp_font_dir}" -name "*.ttf" -exec cp -f {} "${mac_font_dir}/" \; 2>/dev/null || true
            echo "pkt-git: JetBrainsMono Nerd Font installed in ~/Library/Fonts."
        fi
        rm -rf "${tmp_font_dir}"
        return 0
    fi

    # Linux font installation
    local font_dir="${HOME}/.local/share/fonts/NerdFonts"
    if [[ -d "${font_dir}" ]] && ls "${font_dir}"/*JetBrainsMono* >/dev/null 2>&1; then
        return 0
    fi

    echo "pkt-git: Installing JetBrainsMono Nerd Font..."
    mkdir -p "${font_dir}"
    local tmp_font_dir
    tmp_font_dir="$(mktemp -d "${TMPDIR:-/tmp}/pkt-font.XXXXXX")"
    local font_url="https://github.com/ryanoasis/nerd-fonts/releases/latest/download/JetBrainsMono.tar.xz"
    local downloaded=0

    if command -v curl >/dev/null 2>&1; then
        if curl -fsSL "${font_url}" -o "${tmp_font_dir}/font.tar.xz" 2>/dev/null; then
            downloaded=1
        fi
    elif command -v wget >/dev/null 2>&1; then
        if wget -q "${font_url}" -O "${tmp_font_dir}/font.tar.xz" 2>/dev/null; then
            downloaded=1
        fi
    fi

    if [[ "${downloaded}" -eq 1 ]]; then
        tar -xf "${tmp_font_dir}/font.tar.xz" -C "${tmp_font_dir}" 2>/dev/null || true
        find "${tmp_font_dir}" -name "*.ttf" -exec cp -f {} "${font_dir}/" \; 2>/dev/null || true
        if command -v fc-cache >/dev/null 2>&1; then
            fc-cache -f "${font_dir}" >/dev/null 2>&1 || true
        fi
        echo "pkt-git: JetBrainsMono Nerd Font installed successfully."

        # Configure GNOME Terminal font if profile exists
        if command -v gsettings >/dev/null 2>&1; then
            local profile
            profile="$(gsettings get org.gnome.Terminal.ProfilesList default 2>/dev/null | tr -d \')" || true
            if [[ -n "${profile}" ]]; then
                gsettings set "org.gnome.Terminal.Legacy.Profile:/org/gnome/terminal/legacy/profiles:/:${profile}/" use-system-font false 2>/dev/null || true
                gsettings set "org.gnome.Terminal.Legacy.Profile:/org/gnome/terminal/legacy/profiles:/:${profile}/" font 'JetBrainsMono Nerd Font 11' 2>/dev/null || true
            fi
        fi
    else
        echo "note: could not download JetBrainsMono Nerd Font automatically. You can install it manually from https://www.nerdfonts.com" >&2
    fi
    rm -rf "${tmp_font_dir}"
}

# 1. Ensure prerequisites
install_git_if_missing
configure_git_identity_if_needed
install_nerd_font_if_needed

SOURCE="${PKTG_SOURCE:-}"
if [[ -z "${SOURCE}" ]]; then
    if ! SOURCE="$(find_prebuilt)"; then
        if ! command -v go >/dev/null 2>&1; then
            echo "pkt-git: no prebuilt binary found and Go is not installed." >&2
            echo "Build with 'make build' or install a release archive." >&2
            exit 1
        fi
        TMP_BUILD="$(mktemp -d "${TMPDIR:-/tmp}/pkt-git-build.XXXXXX")"
        VERSION="$(git -C "${ROOT_DIR}" describe --tags --always 2>/dev/null || echo "0.1.0")"
        LDFLAGS="-s -w -X github.com/MarkOne-dev/pkt-git/internal/config.Version=${VERSION}"
        (cd "${ROOT_DIR}" && CGO_ENABLED=0 go build -ldflags "${LDFLAGS}" -o "${TMP_BUILD}/pkt" ./cmd/pkt)
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

configure_path_if_needed

echo ""
"${TARGET}" welcome || true

echo ""
echo "pkt-git successfully installed at ${TARGET}."
