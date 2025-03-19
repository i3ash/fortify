# Introduction

**Fortify** is a command-line tool designed to enhance file security through encryption. It is available as a standalone static executable binary and can be installed on a variety of platforms.

## Main Features

- 🔒 Encrypt files with AES-256 encryption.
- 🛡️ Protect AES secret keys using Shamir's Secret Sharing (SSS) or RSA encryption.
- 🔄 Support for both encryption and decryption.
- 🚀 Execute encrypted files directly using specified keys.

## Quick Start Guide

`fortify` is available as a standalone static executable binary and doesn't require any external dependencies.

### Installation Steps

- **Linux**:  
  ```shell
  /usr/bin/env sh -c "$(curl -fsSL https://i3ash.com/fortify/install.sh)"
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
- **Download**: Precompiled binaries are available [here](https://github.com/i3ash/fortify/releases).

For more in-depth instructions on using Fortify, visit the [Getting Started](/getting-started.md) guide.