# Introduction

`fortify` is a command-line security tool for file encryption and protection.

## Main Features

- 📦 Golang implementation of Shamir’s Secret Sharing (SSS).
- 🧩 Split secret files into multiple shares with SSS, and reconstruct files when the threshold number of shares is
  available.
- 🔒 Encrypt files using AES-256, protecting AES keys with either SSS or RSA encryption.
- 🚀 Directly execute encrypted files by providing the required keys.

## Installation

`fortify` is distributed as a standalone static executable with zero external dependencies.
Download pre-compiled binaries from our [GitHub Releases](https://github.com/i3ash/fortify/releases) page, or install
using one of these methods:

- **Linux**:
  ```shell
  /usr/bin/env sh -c "$(curl -fsSL https://fortify.i3ash.com/install.sh)"
  ```
- **macOS**:
  ```shell
  brew install i3ash/bin/fortify
  ```
- **Go Install**:
  ```shell
  go install github.com/i3ash/fortify@latest
  ```
- **Docker**:
  ```shell
  docker run --rm i3ash/fortify version
  ```

For more detailed instructions and usage examples, please refer to the [Getting Started](getting-started.md) guide.
