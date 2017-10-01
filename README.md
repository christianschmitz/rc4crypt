# rc4crypt

## Synopsis

Simple command line tool to encrypt/decrypt plain text. Exactly the same algorithm as the TXTcrypt app by Vlad Alexa: vladalexa.com/apps/ios/txtcrypt

You are prompted for a pass-phrase. Internally this is converted to a 256-byte key. The RC4 encryption algorithm is used.

## Usage

`rc4crypt [-p | -s <suffix> [<file> ..[<file>]]] [-h]`

Options:

* `-p`: print the key generated from the pass-phrase. This key can then be used in other encryption tools like openssl or mcrypt.
* `-s <suffix>`: generate output filenames by appending this suffix to input filenames.  Ignored when reading from stdin.
* `-h`: Print the help message.

Details:

* Read from stdin if no files are specified.
* Output is printed to stdout when reading from stdin.
* Output is printed to stdout if no suffix is specified for input files.
* Pass-phrase is read from terminal.
* Decrypt by entering blank pass-phrase at second prompt.
* Uses rc4 (a.k.a. arcfour) algorithm and base64 encoding.
* Encrypts/decrypts plain text into plain text.

## Installation

Compile:
`go build rc4crypt.go`

Copy the binary into a directory in your PATH.
