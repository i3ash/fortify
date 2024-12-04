#!/usr/bin/env bash
# shellcheck disable=SC2317
set -eou pipefail

source .cidoer/cidoer.core.sh

declare -rx ARTIFACT_CMD='fortify'

define_custom_test() {
  test_custom_do() {
    go test -v ./...
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
  do_print_dash_pair "ARTIFACT_VERSION" "$($out version)"
}

define_custom_prepare() {
  prepare_custom_init() {
    check_go
  }
  prepare_custom_do() {
    local tag
    tag=$(custom_version_tag)
    export ARTIFACT_TAG="$tag"
    do_print_dash_pair 'ARTIFACT_TAG' "$ARTIFACT_TAG"
    sed -e "s|#VERSION|${tag}|g" < "cmd/version.go-e" > "cmd/version.go"
  }
  custom_version_tag() {
    local tag count hash
    tag=$(do_git_version_tag)
    count=$(do_git_count_commits_since "$tag")
    hash=$(do_git_short_commit_hash)
    printf '%s+%s-%s' "$tag" "$count" "$hash"
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

define_custom_build() {
  build_custom_init() {
    local system architecture
    system=$(uname -s | tr '[:upper:]' '[:lower:]')
    architecture=$(uname -m | tr '[:upper:]' '[:lower:]' | sed 's/amd64/x86_64/g')
    local filename="${ARTIFACT_CMD}-${system}-${architecture}"
    export ARTIFACT_FILENAME="$filename"
    do_print_dash_pair 'ARTIFACT_FILENAME' "$ARTIFACT_FILENAME"
  }
  build_custom_do() {
    local out="${OUT_DIR:-.}/${ARTIFACT_FILENAME}"
    CGO_ENABLED=0 go build -a -installsuffix nocgo -ldflags '-s -w' -o "$out"
    file "$out"
    do_print_dash_pair "ARTIFACT_VERSION" "$($out version)"
    ln -s "$out" "${OUT_DIR:-.}/${ARTIFACT_CMD}"
  }
}

define_custom_build_darwin_universal() {
  build_darwin_universal_custom_init() {
    export ARTIFACT_FILENAME="${ARTIFACT_CMD}-darwin-universal"
    do_print_dash_pair 'ARTIFACT_FILENAME' "$ARTIFACT_FILENAME"
  }
  build_darwin_universal_custom_do() {
    local out="${OUT_DIR:-.}/${ARTIFACT_FILENAME}"
    lipo -create -output "$out" "${OUT_DIR:-.}/${ARTIFACT_CMD}-darwin-x86_64" "${OUT_DIR:-.}/${ARTIFACT_CMD}-darwin-arm64"
    file "$out"
    do_print_dash_pair "ARTIFACT_VERSION" "$($out version)"
  }
}

define_custom_build_darwin_arm64() {
  build_darwin_arm64_custom_do() {
    build_artifact darwin arm64
  }
}

define_custom_build_darwin_x64() {
  build_darwin_x64_custom_do() {
    export ARTIFACT_FILENAME="$ARTIFACT_CMD-darwin-x86_64"
    build_artifact darwin amd64
  }
}

define_custom_build_linux_x64() {
  build_linux_x64_custom_do() {
    export ARTIFACT_FILENAME="$ARTIFACT_CMD-linux-x86_64"
    build_artifact linux amd64
  }
}

define_custom_build_linux_aarch64() {
  build_linux_aarch64_custom_do() {
    export ARTIFACT_FILENAME="$ARTIFACT_CMD-linux-aarch64"
    build_artifact linux arm64
  }
}

define_custom_build_linux_riscv64() {
  build_linux_riscv64_custom_do() {
    build_artifact linux riscv64
  }
}

define_custom_build_linux_mips64le() {
  build_linux_mips64le_custom_do() {
    build_artifact linux mips64le
  }
}

define_custom_build_windows_x64() {
  build_windows_x64_custom_do() {
    export ARTIFACT_FILENAME="$ARTIFACT_CMD-windows-x86_64"
    build_artifact windows amd64
  }
}
