# Introduction to Fortify 🌐

Welcome to **Fortify**, a command-line tool designed for file security through encryption. Fortify uses advanced cryptographic methods to keep your files safe from unauthorized access.

## Key Features 🔑

- **File Encryption**: Encrypt files using the AES-256 standard.
- **Key Protection**: Secure AES secret keys with Shamir's Secret Sharing (SSS) or RSA encryption methods.
- **Flexible Decryption**: Decrypt or execute protected files as needed.

## Installation 

`fortify` is available as a standalone static executable binary without additional dependencies. Here's how you can install it:

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

- **Direct Download**: Get the precompiled binary from [releases](https://github.com/i3ash/fortify/releases).

## Using Fortify

Fortify offers two primary encryption methods:

### Shamir's Secret Sharing (SSS) 🔐

- **Encryption**: 
  Use randomly generated key parts:
  ```shell
  fortify encrypt -i <input_file> -o <output_file>
  ```

- **Decryption**:
  ```shell
  fortify decrypt -i <fortified_file> <key_part1> <key_part2> ...
  ```

- **Execution**:
  ```shell
  fortify execute -i <fortified_file> <key_part1> <key_part2> ...
  ```

### RSA Encryption

- **Encryption**: 
  ```shell
  fortify encrypt -i <input_file> -k rsa <public_key_file>
  ```

- **Decryption**: 
  ```shell
  fortify decrypt -i <fortified_file> <private_key_file>
  ```

- **Execution**: 
  ```shell
  fortify execute -i <fortified_file> <private_key_file>
  ```

For detailed development instructions, licensing, and contribution guidelines, check out the [Developer's Guide](https://github.com/i3ash/fortify/blob/main/README_DEV.md). Fortify aims to securely protect your files. 🛡️