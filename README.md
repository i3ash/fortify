# Fortify

`fortify` is a command-line security tool for file encryption and protection.

## Main Features

- 📦 Golang implementation of Shamir’s Secret Sharing (SSS).
- 🧩 Split secret files into multiple shares with SSS, and reconstruct files when the threshold number of shares is
  available.
- 🔒 Encrypt files using AES-256, protecting AES keys with either SSS or RSA encryption.
- 🚀 Directly execute encrypted files by providing the required keys.

## User Guide

- [Installation](https://fortify.i3ash.com)
- [Getting Started](https://fortify.i3ash.com/getting-started.html)

---

# Developer Guide

[![Release](https://github.com/i3ash/fortify/actions/workflows/release.yml/badge.svg)](https://github.com/i3ash/fortify/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/i3ash/fortify)](https://goreportcard.com/report/github.com/i3ash/fortify)

## License

This project is licensed under the MIT License.

## Contributing

We welcome contributions through issue submissions and pull requests. Feel free to suggest improvements or report
issues.

## Build

To build the project, run:

```shell
bash build.sh
```

After building, execute the following commands to confirm the result:
```shell
./build/fortify version -d
```
```shell
./build/fortify help
```

>>> [Check out more details](README_DEV.md)
