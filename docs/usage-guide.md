# Usage Guide 🛡️

Fortify is a command-line tool designed to enhance file security through encryption, decryption, and execution using Shamir's Secret Sharing (SSS) and RSA encryption. This guide provides a walkthrough on how to utilize Fortify for your encryption needs.

## Getting Started 🚀

Ensure Fortify is installed through one of the following methods: Linux, macOS, Go, Docker, or by downloading a standalone binary.

### Using Shamir's Secret Sharing (SSS) Method

#### Encryption

Secure your file by encrypting with AES-256 and SSS:

- For random key shares:

  ```shell
  fortify sss random -b 32 -p <number_of_shares> -t <threshold>
  fortify encrypt -i <input_file> -o <output_file>
  ```

- For specific key shares:

  ```shell
  fortify encrypt -i <input_file> <key_part1> <key_part2> ...
  ```

#### Decryption

Retrieve your original file with:

```shell
fortify decrypt -i <fortified_file> <key_part1> <key_part2> ...
```

#### Execution

Run your encrypted file seamlessly:

```shell
fortify execute -i <fortified_file> <key_part1> <key_part2> ...
```

### Utilizing RSA

#### Encryption

Use your RSA public key for encryption:

```shell
fortify encrypt -i <input_file> -k rsa <public_key_file>
```

#### Decryption

Decipher the file using your private key:

```shell
fortify decrypt -i <fortified_file> <private_key_file>
```

#### Execution

Execute using your RSA private key:

```shell
fortify execute -i <fortified_file> <private_key_file>
```

