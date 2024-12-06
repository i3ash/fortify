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

## Shamir's Secret Sharing (SSS)

### Splitting and Combining Secret Shares

To split and combine secret shares, use the following commands:

```shell
pushd build/sss && ../fortify sss split -vT ../fortify; popd
```

```shell
pushd build/sss && ../fortify sss combine -o combined.out -vT 1of5.json 3of5.json 5of5.json; popd
```

**Tips:**

- For enhanced security, store generated secret shares in different locations.
- While suitable for processing large files, this method may not be optimal for smaller files.

### Encrypting with Randomly Generated Secret Key

Encrypt files with randomly generated key parts:

```shell
pushd build/sss && ../fortify encrypt -i ../fortify -T; popd
```

Decrypt fortified files with specified key parts:

```shell
pushd build/sss && ../fortify decrypt -i fortified.data -T fortified.key1of2.json fortified.key2of2.json; popd
```

Execute fortified files with specified key parts:

```shell
pushd build/sss && ../fortify execute -i fortified.data fortified.key1of2.json fortified.key2of2.json -- encrypt -h; popd
```

### Encrypting and Decrypting with Specified Key Parts

Generate new random key parts:

```shell
pushd build/sss && ../fortify sss random -p3 -t2 --prefix p; popd
```

Encrypt files using specified key parts:

```shell
pushd build/sss && ../fortify encrypt -i ../fortify -vT p1of3.json p2of3.json; popd
```

Decrypt fortified files using specified key parts:

```shell
pushd build/sss && ../fortify decrypt -i fortified.data -vT p1of3.json p3of3.json; popd
```

Execute fortified files using specified key parts:

```shell
pushd build/sss && ../fortify execute -i fortified.data p2of3.json p3of3.json; popd
```

## RSA Encryption

Generate RSA key pairs:

```shell
bash debug_keygen.sh
```

### Encrypting with RSA Public Key

Encrypt files using RSA public key:

```shell
pushd build/rsa && ../fortify encrypt -i ../fortify -k rsa -vT ../../debug/key_rsa/id_rsa.pub; popd
```

Encrypt files using RSA public key in PEM format:

```shell
pushd build/rsa && ../fortify encrypt -i ../fortify -k rsa -vT ../../debug/key_rsa/id_rsa_pem.pub; popd
```

```shell
pushd build/rsa && ../fortify encrypt -i ../fortify -k rsa  -vT ../../debug/key_rsa/id_rsa_pkcs8.pub; popd
# Will Fail
```

> - PKCS #8 public key is unsupported

```shell
pushd build/rsa && ../fortify encrypt -i ../fortify -k rsa -vT ../../debug/key_rsa/id_rsa_rfc4716.pub; popd
# Will Fail
```

> - RFC 4716 public key is unsupported

### Execute Fortified Files with RSA Private Key

Execute fortified files using RSA private key:

```shell
pushd build/rsa && ../fortify execute -i fortified.data ../../debug/key_rsa/id_rsa; popd
```

Execute fortified files using RSA private key in PEM format:

```shell
pushd build/rsa && ../fortify execute -i fortified.data ../../debug/key_rsa/id_rsa_pem; popd
```

Execute fortified files using RSA private key in RFC 4716 format:

```shell
pushd build/rsa && ../fortify execute -i fortified.data ../../debug/key_rsa/id_rsa_rfc4716; popd
```

```shell
pushd build/rsa && ../fortify execute -i fortified.data ../../debug/key_rsa/id_rsa_pkcs8; popd
# Will Fail
```

> - encrypted PKCS #8 private key is unsupported

---