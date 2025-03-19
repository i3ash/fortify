# Using Fortify

Fortify offers two primary encryption methods:

## 1. Run `fortify` with SSS (Shamir's Secret Sharing)

### Encryption

Encrypt files with randomly generated key parts:

`fortify encrypt -i <input_file> -o <output_file>`

Or generate key parts first

`fortify sss random -b 32 -p <number_of_shares> -t <threshold>`

then encrypt files with specified key parts:

`fortify encrypt -i <input_file> <key_part1> <key_part2> ...`

### Decryption

Decrypt files with specified key parts:

`fortify decrypt -i <fortified_file> <key_part1> <key_part2> ...`

### Execution

Execute fortified files with specified key parts:

`fortify execute -i <fortified_file> <key_part1> <key_part2> ...`


## 2. Run `fortify` with RSA

### Encryption

Encrypt files with RSA public key:

`fortify encrypt -i <input_file> -k rsa <public_key_file>`

### Decryption

Decrypt files with RSA private key:

`fortify decrypt -i <fortified_file> <private_key_file>`

### Execution

Execute fortified files with RSA private key:

`fortify execute -i <fortified_file> <private_key_file>`

For detailed development instructions, licensing, and contribution guidelines, check out the [Developer's Guide](https://github.com/i3ash/fortify/blob/main/README_DEV.md).
