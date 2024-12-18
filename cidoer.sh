#!/usr/bin/env bash
# shellcheck disable=SC2317
set -eou pipefail

if [ ! -f .cidoer/cidoer.core.sh ]; then
  /usr/bin/env sh -c "$(curl -fsSL https://i3ash.com/cidoer/install.sh)" -- '1.0.4'
fi
source .cidoer/cidoer.core.sh

declare -rx ARTIFACT_CMD='fortify'

define_test() {
  test_do() {
    go test -v ./...
  }
}

define_prepare() {
  prepare_do() {
    local tag
    tag=$(custom_version_tag)
    export ARTIFACT_TAG="$tag"
    do_print_dash_pair 'ARTIFACT_TAG' "$ARTIFACT_TAG"
    do_print_dash_pair 'DO_NOT_REPLACE_VERSION' "${DO_NOT_REPLACE_VERSION:-}"
    if [ 'yes' != "${DO_NOT_REPLACE_VERSION:-}" ]; then
      do_replace \< \> <cmd/version.go-e >cmd/version.go
    fi
    check_go
  }
  custom_version_tag() {
    local tag count hash
    tag=$(do_git_version_tag)
    count=$(do_git_count_commits_since "$tag")
    hash=$(do_git_short_commit_hash)
    if [ "$count" -gt 0 ]; then
      printf '%s%s%s' "$tag" ".$count" "-$hash"
    else printf '%s' "$tag"; fi
  }
  check_go() {
    if ! command -v go &>/dev/null; then
      echo "Error: Go language environment not found." >&2
      exit 1
    fi
    printf "Check: " >&1
    which go >&1
  }
}

build_artifact() {
  local os="${1:?Require os type}"
  local arch="${2:?Require architecture}"
  export ARTIFACT_FILENAME="${ARTIFACT_FILENAME:-$(printf '%s' "$ARTIFACT_CMD-$os-$arch")}"
  do_print_dash_pair 'ARTIFACT_FILENAME' "${ARTIFACT_FILENAME}"
  local out="${OUT_DIR:-.}/${ARTIFACT_FILENAME}"
  GOARCH=$arch GOOS=$os CGO_ENABLED=0 go build -a -installsuffix nocgo -ldflags '-s -w' -o "$out"
  file "$out"
  do_print_dash_pair 'ARTIFACT_OUT' "$out"
  if [ 'darwin' = "$(do_os_type)" ] || [[ "${ARTIFACT_FILENAME:-}" == *"$(do_host_type)"* ]]; then
    do_print_dash_pair "ARTIFACT_VERSION" "$($out version)"
  fi
}

define_build_linux_x64() {
  build_linux_x64_do() {
    export ARTIFACT_FILENAME="$ARTIFACT_CMD-linux-x86_64"
    build_artifact linux amd64
  }
}

define_build_linux_arm64() {
  build_linux_arm64_do() {
    build_artifact linux arm64
  }
}

define_build_linux_riscv64() {
  build_linux_riscv64_do() {
    build_artifact linux riscv64
  }
}

define_build_linux_mips64le() {
  build_linux_mips64le_do() {
    build_artifact linux mips64le
  }
}

define_build_windows_x64() {
  build_windows_x64_do() {
    export ARTIFACT_FILENAME="$ARTIFACT_CMD-windows-x86_64.exe"
    build_artifact windows amd64
  }
}

define_build_windows_arm64() {
  build_windows_arm64_do() {
    export ARTIFACT_FILENAME="$ARTIFACT_CMD-windows-arm64.exe"
    build_artifact windows arm64
  }
}

define_build_darwin_arm64() {
  build_darwin_arm64_do() {
    build_artifact darwin arm64
  }
}

define_build_darwin_x64() {
  build_darwin_x64_do() {
    export ARTIFACT_FILENAME="$ARTIFACT_CMD-darwin-x86_64"
    build_artifact darwin amd64
  }
}

define_build_darwin_universal() {
  build_darwin_universal_do() {
    export ARTIFACT_FILENAME="${ARTIFACT_CMD}-darwin-universal"
    do_print_dash_pair 'ARTIFACT_FILENAME' "$ARTIFACT_FILENAME"
    local out="${OUT_DIR:-.}/${ARTIFACT_FILENAME}"
    local cmd1="${ARTIFACT_DARWIN_X64:-$(printf '%s' "${OUT_DIR:-.}/${ARTIFACT_CMD}-darwin-x86_64")}"
    local cmd2="${ARTIFACT_DARWIN_A64:-$(printf '%s' "${OUT_DIR:-.}/${ARTIFACT_CMD}-darwin-arm64")}"
    lipo -create -output "$out" "$cmd1" "$cmd2"
    file "$out"
    chmod +x "$out"
    do_print_dash_pair 'ARTIFACT_OUT' "$out"
    do_print_dash_pair "ARTIFACT_VERSION" "$($out version)"
  }
}

define_docker_debian() {
  docker_debian_do() {
    local tags=(--tag "$DOCKER_IMAGE:$ARTIFACT_TAG")
    if [ "$LATEST_TAG" = "true" ]; then
      tags+=(--tag "$DOCKER_IMAGE:debian")
    fi
    docker buildx build "${tags[@]}" \
      --platform linux/386,linux/amd64,linux/arm/v5,linux/arm/v7,linux/arm64/v8,linux/mips64le,linux/ppc64le,linux/s390x \
      --target debian --push .
  }
}

define_docker_alpine() {
  docker_alpine_do() {
    local tags=(--tag "$DOCKER_IMAGE:$ARTIFACT_TAG")
    if [ "$LATEST_TAG" = "true" ]; then
      tags+=(--tag "$DOCKER_IMAGE:alpine")
    fi
    docker buildx build "${tags[@]}" \
      --platform linux/386,linux/amd64,linux/arm/v6,linux/arm/v7,linux/arm64/v8,linux/ppc64le,linux/riscv64,linux/s390x \
      --target alpine --push .
  }
}

define_docker_busybox() {
  docker_busybox_do() {
    local tags=(--tag "$DOCKER_IMAGE:$ARTIFACT_TAG")
    if [ "$LATEST_TAG" = "true" ]; then
      tags+=(--tag "$DOCKER_IMAGE:busybox")
    fi
    docker buildx build "${tags[@]}" \
      --platform linux/386,linux/amd64,linux/arm/v5,linux/arm/v7,linux/arm64/v8,linux/mips64le,linux/ppc64le,linux/riscv64,linux/s390x \
      --target busybox --push .
  }
}

define_docker_minimal() {
  docker_minimal_do() {
    local tags=(--tag "$DOCKER_IMAGE:$ARTIFACT_TAG")
    if [ "$LATEST_TAG" = "true" ]; then
      tags+=(--tag "$DOCKER_IMAGE:latest")
    fi
    docker buildx build "${tags[@]}" \
      --platform linux/386,linux/amd64,linux/arm/v5,linux/arm/v6,linux/arm/v7,linux/arm64/v8,linux/mips64le,linux/ppc64le,linux/riscv64,linux/s390x \
      --target minimal --push .
  }
}
