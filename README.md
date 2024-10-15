# Fortify

**Fortify** is a command-line tool designed to enhance file security through encryption.

## Features

* Fortifies any file through encryption, then decrypts or executes the fortified file.
* Encrypts file using AES-256.
* Protects the AES secret key with either Shamir's Secret Sharing (SSS) or RSA encryption.

## Usage Overview

### Installing `fortify`

`fortify` is distributed as a standalone static executable binary, requiring no external dependencies.

#### 1. Linux

```shell
/usr/bin/env sh -c "$(curl -fsSL https://i3ash.com/fortify/install.sh)"
```

#### 2. macOS

```shell
brew install i3ash/bin/fortify
```

#### 3. go install

```shell
go install github.com/i3ash/fortify@latest
```

#### 4. Docker

```shell
docker run --rm i3ash/fortify version
```

#### 5. Download

Download a precompiled binary from [here](https://github.com/i3ash/fortify/releases).


### Run `fortify` with SSS (Shamir's Secret Sharing)

#### Encryption

Encrypt files with randomly generated key parts:

`fortify encrypt -i <input_file> -o <output_file>`

Encrypt files with specified key parts:

`fortify sss random -b 32 -p <number_of_shares> -t <threshold>`

`fortify encrypt -i <input_file> <key_part1> <key_part2> ...`

#### Decryption

Decrypt files with specified key parts:

`fortify decrypt -i <fortified_file> <key_part1> <key_part2> ...`

#### Execution

Execute fortified files with specified key parts:

`fortify execute -i <fortified_file> <key_part1> <key_part2> ...`

### Run `fortify` with RSA

#### Encryption

Encrypt files with RSA public key:

`fortify encrypt -i <input_file> -k rsa <public_key_file>`

#### Decryption

Decrypt files with RSA private key:

`fortify decrypt -i <fortified_file> <private_key_file>`

#### Execution

Execute fortified files with RSA private key:

`fortify execute -i <fortified_file> <private_key_file>`


---

# Developer's Guide

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
pushd build && ./fortify -h && ./fortify version; popd
```

>>> [Check out more details](https://github.com/i3ash/fortify/blob/main/README_DEV.md)
