# Encrypting Text Files with Fortify

## Introduction

This guide will show you how to encrypt text files using Fortify. It will also detail how to install the binary. You'll be encrypting your files in no time.

Let's get started!

## Prerequisites

Make sure the Fortify library is installed on your system. Choose your platform and follow the installation instructions below:

### Installation

#### For Linux

```shell
/usr/bin/env sh -c "$(curl -fsSL https://i3ash.com/fortify/install.sh)"
```

#### For macOS

```shell
brew install i3ash/bin/fortify
```

#### Using `go install`

```shell
go install github.com/i3ash/fortify@latest
```

#### Using Docker

```shell
docker run --rm i3ash/fortify version
```

#### Download Binary

Download the binary from [here](https://github.com/i3ash/fortify/releases).


## Step 1: Encrypting Text Files with SSS 🛡️

### Generate Key Parts

First, create key parts using Shamir's Secret Sharing. These keys are necessary for decrypting the files.

```shell
fortify sss random -b 32 -p <number_of_shares> -t <threshold>
```

- `<number_of_shares>`: Number of key parts to generate.
- `<threshold>`: Minimum number of key parts needed to decrypt.

### Encrypt Your Text File

To encrypt a text file using the generated keys, use:

```shell
fortify encrypt -i <input_file> -o <output_file> <key_part1> <key_part2> ...
```

## Step 2: Decrypting Text Files

Decrypt your file by providing the necessary key parts:

```shell
fortify decrypt -i <output_file> <key_part1> <key_part2> ...
```

## Step 3: Encrypting with RSA 🔒

If you would like to use RSA encryption instead, follow these steps:

### Encrypt Your Text File with RSA

Use a public key for encryption:

```shell
fortify encrypt -i <input_file> -k rsa <public_key_file>
```

### Decrypt Your Text File

Use your private key to decrypt the file:

```shell
fortify decrypt -i <output_file> <private_key_file>
```

## Conclusion

Fortify provides you with a secure way to store text, using advanced cryptographic techniques. This system ensures that only authorized users can access sensitive information.

In this guide, we showcased how to install Fortify and  various commands available for encrypting and decrypting files.