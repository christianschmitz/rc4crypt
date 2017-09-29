# rc4crypt

## Synopsis

Simple command line tool to encrypt/decrypt plain text. Exactly the same algorithm as the TXTcrypt app by Vlad Alexa: vladalexa.com/apps/ios/txtcrypt

You are prompted for a pass-phrase. Internally this is converted to a 256-byte key. The RC4 encryption algorithm is used.

## Usage

`rc4crypt [OPTIONS] [FILE1 [FILE2 ...]]`

Options:
* `-d` : decrypt
* `-p` : printkey
* `-s SUFFIX` : FILE1 -> FILE1SUFFIX, useful for batch processing many files and only having to enter pass-phrase once

If no files are specified input is read from stdin. Output from encrypting/decrypting stdin is always sent to stdout. Output from encrypting/decrypting files is sent to stdout unless a suffix is specified.

## Installation

Compile:
`go build rc4crypt.go`

Copy the binary into a directory in your PATH.
