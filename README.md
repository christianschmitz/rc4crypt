# rc4crypt

## Synopsis

Simple command line tool to encrypt/decrypt plain text. Exactly the same algorithm as the TXTcrypt app by Vlad Alexa: vladalexa.com/apps/ios/txtcrypt

You are prompted for a pass-phrase. Internally this is converted to a 256-byte key. The rc4 encryption algorithm is used.

## Usage

`rc4crypt [-p | [-h] [-s <suffix>] [<file> [<file> ..]]]`

Options:

* `-p`: print the key generated from the pass-phrase. This key can then be used in other crypto tools (e.g. openssl or mcrypt). Prints to stdout.
* `-s <suffix>`: generate output filename(s) by appending this suffix to input filename(s).  Ignored when reading from stdin.
* `-h`: print the help message.

Details:

* Read from stdin if no files are specified.
* Output is printed to stdout when reading from stdin.
* Output is printed to stdout if no suffix is specified for input file(s).
* Pass-phrase is read from terminal.
* Decrypt by entering blank pass-phrase at second prompt.
* Uses rc4 (a.k.a. arcfour) algorithm and base64 encoding.
* Encrypts/decrypts plain text into plain text.

## Installation

Download:
```
git clone https://github.com/christianschmitz/rc4crypt ./rc4crypt
cd rc4crypt
```

Compile:
```
go build rc4crypt.go
```

Copy the binary into a directory in your PATH. E.g:
```
echo 'export PATH=${PATH}:~/bin' >> ~/.bashrc
cp rc4crypt ~/bin/
```

Open a new terminal and try:
```
rc4crypt -h
```

## Dependencies

* linux system
* git toolchain
* golang toolchain
* golang-golang-x-crypto-dev package

## Copyright

Christian Schmitz

## Contact

christian.schmitz@telenet.be
