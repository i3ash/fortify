# Installation Guide

Welcome to the **Fortify Installation Guide**. This guide provides instructions for installing Fortify on various platforms, helping you enhance your file security. 🛡️

## Platforms Supported

Fortify is a versatile tool available on multiple platforms via a standalone static executable binary that requires no external dependencies. 🎉

### Linux Installation

To install Fortify on Linux, run the following command in your terminal:

```shell
/usr/bin/env sh -c "$(curl -fsSL https://i3ash.com/fortify/install.sh)"
```

### macOS Installation

For macOS users, you can install Fortify using the Homebrew package manager:

```shell
brew install i3ash/bin/fortify
```

### Go Installation

Alternatively, if you have a Go environment set up, install Fortify with:

```shell
go install github.com/i3ash/fortify@latest
```

### Docker Installation

If you prefer using Docker, you can run Fortify with:

```shell
docker run --rm i3ash/fortify version
```

### Direct Download

For a direct setup, download the precompiled binaries [here](https://github.com/i3ash/fortify/releases).

Fortify installs efficiently and offers robust encryption through AES-256, RSA, or Shamir's Secret Sharing. Customize your security needs to keep your data safe. 🔐

For more details on using Fortify, including encryption and decryption instructions, refer to the [Developer's Guide](https://github.com/i3ash/fortify/blob/main/README_DEV.md).