#!/usr/bin/env sh
set -eu

artifact="fortify"
github_user="i3ash"

check_installed() {
  cmd="$1"
  cmd_path=$(command -v "${cmd:?}" 2>/dev/null)
  if [ -n "${cmd_path}" ] && [ -x "${cmd_path}" ]; then
    printf "%-8s is already installed at path: %s\n" "${cmd}" "${cmd_path}" >&2
    return 0
  fi
  return 1
}

check_dependencies() {
  required_commands="${1}"
  optional_commands="${2}"
  missing=0
  for cmd in $required_commands; do
    if ! check_installed "$cmd"; then
      echo "Error: Required command '$cmd' is not executable or not in PATH." >&2
      missing=1
    fi
  done
  for cmd in $optional_commands; do
    if ! check_installed "$cmd"; then
      echo "Warning: Optional command '$cmd' is not executable or not in PATH. Some functionalities may be limited." >&2
    fi
  done
  if [ "$missing" -eq 1 ]; then
    echo "Please install the missing dependencies and try again." >&2
    exit 1
  fi
  unset required_commands
  unset optional_commands
  unset missing
  unset cmd
  unset cmd_path
}

echo_version_stable() {
  url="https://raw.githubusercontent.com/${github_user}/${artifact}/refs/heads/main/stable.txt"
  if ! response=$(curl -sSL --fail --retry 3 --retry-delay 2 "${url:-}"); then
    echo "Failed to fetch stable version after multiple attempts." >&2
    echo ""
  else
    if printf '%s' "$response" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
      echo "$response"
    else
      echo "Invalid version format received." >&2
      echo ""
    fi
  fi
  unset url
  unset response
}

usage() {
  echo "Usage: $0 [version]"
  echo "Download and install the specified version of the (${artifact}) command."
}

detect_arch() {
  arch=''
  if [ -n "${HOSTTYPE:-}" ]; then
    arch="${HOSTTYPE:-}"
  else
    if command -v uname >/dev/null 2>&1; then arch="$(uname -m)"; fi
  fi
  arch="$(printf '%s' "${arch:-}" | tr '[:upper:]' '[:lower:]')"
  case "$arch" in
  x86_64 | amd64 | x64) type='x86_64' ;;
  i*86 | x86) type='x86' ;;
  arm64 | aarch64) type='arm64' ;;
  armv5* | armv6* | armv7* | aarch32) type='arm' ;;
  armv8*) type="$arch" ;;
  ppc | powerpc) type='ppc' ;;
  ppc64 | ppc64le) type="$arch" ;;
  mips | mips64 | mipsle | mips64le | s390x | riscv64) type="$arch" ;;
  *) type='unknown' ;;
  esac
  printf '%s' "$type"
  unset arch
  unset type
}

detect_os() {
  if [ -n "${OSTYPE:-}" ]; then
    system="${OSTYPE:-}"
  else
    if command -v uname >/dev/null 2>&1; then system="$(uname -s)"; fi
  fi
  system="$(printf '%s' "${system:-}" | tr '[:upper:]' '[:lower:]')"
  case "$system" in
  linux*) system='linux' ;;
  darwin*) system='darwin' ;;
  cygwin* | msys* | mingw* | windows*) system='windows' ;;
  *) system='unknown' ;;
  esac
  printf '%s' "$system"
  unset system
}

move_file() {
  src="$1"
  dst="$2"
  if ! mv "${src}" "${dst}"; then
    echo "Failed to move file to ${dst}" >&2
    return 1
  fi
  unset src
  unset dst
}

install_asset() {
  tmp_file="$(mktemp)"
  os="$(detect_os)" || exit 1
  arch="$(detect_arch)" || exit 1
  asset="${artifact}-${os}-${arch}"
  url="https://github.com/${github_user}/${artifact}/releases/download/${version}/${asset}"
  echo "Downloading: ${url}"
  if ! curl -sSLf --retry 3 --retry-delay 2 "${url}" -o "${tmp_file}"; then
    echo "Failed to download ${url} after multiple attempts." >&2
    cleanup
    exit 1
  fi
  install_path="/usr/local/bin"
  if [ ! -w "$install_path" ]; then
    echo "$install_path is not writable."
    echo "Attempting to use sudo..."
    if sudo -n true 2>/dev/null || sudo -v 2>/dev/null; then
      if ! sudo mv "${tmp_file}" "$install_path/$artifact"; then
        echo "Failed to move file to $install_path/$artifact" >&2
        cleanup
        exit 1
      fi
      if ! sudo chmod +x "$install_path/$artifact"; then
        echo "Failed to set execute permission on $install_path/$artifact" >&2
        cleanup
        exit 1
      fi
    else
      echo "sudo not available or permission denied."
      install_path="$HOME/bin"
      mkdir -p "$install_path"
      move_file "${tmp_file}" "$install_path/$artifact"
      if ! chmod +x "$install_path/$artifact"; then
        echo "Failed to set execute permission on $install_path/$artifact" >&2
        cleanup
        exit 1
      fi
      echo "Installed to $install_path/$artifact. Please ensure this directory is in your PATH."
      if ! echo ":$PATH:" | grep -q ":$HOME/bin:"; then
        echo "Note: $HOME/bin is not in your PATH."
        echo 'You can add it to your PATH by running:'
        # shellcheck disable=SC2016
        echo 'export PATH="$HOME/bin:$PATH"'
      fi
    fi
  else
    move_file "${tmp_file}" "$install_path/$artifact"
    if ! chmod +x "$install_path/$artifact"; then
      echo "Failed to set execute permission on $install_path/$artifact" >&2
      cleanup
      exit 1
    fi
  fi
  echo "$install_path/$artifact version"
  "$install_path/$artifact" version
  echo "Finished"
  unset tmp_file
  unset os
  unset arch
  unset asset
  unset url
  unset install_path
}

cleanup() {
  [ -f "${tmp_file:-}" ] && rm -f "${tmp_file}"
  return 0
}
trap cleanup EXIT

check_dependencies 'chmod mv rm mkdir mktemp uname tr curl grep' 'sudo'
if check_installed ${artifact}; then
  echo "Installed version: $(${artifact} version)" >&2
fi

version_stable="$(echo_version_stable)"
version="${1:-${version_stable:?}}"
if [ $# -gt 1 ]; then
  echo "Installation of version '${version}' cancelled because only one argument is allowed." >&2
  usage >&2
  exit 1
fi

install_asset
