package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
)

func readStdin() []byte {
	r := bufio.NewReader(os.Stdin)
	bytes, _ := r.ReadBytes('\x00')

	return bytes
}

func readFile(fname string) []byte {
	var bytes []byte

	if fname == "stdin" {
		bytes = readStdin()
	} else {
		f, err := os.Open(fname)
		if err != nil {
			log.Fatal(err)
		}

		r := bufio.NewReader(f)
		bytes, _ = r.ReadBytes('\x00')
	}

	return bytes
}

func parseArgs() (decrypt bool, printKey bool, suffix string, fnames []string) {
	d := flag.Bool("d", false, "decrypt")
	p := flag.Bool("p", false, "print key")
	s := flag.String("s", "", "suffix (ignored when if no files are specified)")

	var usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: rc4crypt [options] [FILE1 [FILE2 ..]] \noptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "output is printed stdout, pass-phrase is read from /dev/tty (or stdin as backup)\n")
		fmt.Fprintf(os.Stderr, "if no files are specified stream is read from stdin\n")
	}

	flag.Usage = usage

	flag.Parse()

	fnames = flag.Args()

	decrypt = *d
	printKey = *p
	suffix = *s

	return
}

func readPassPhrase(decrypt bool) []byte {
	f, err := os.Open("/dev/tty")
	if err != nil {
		f = os.Stdin
	}

	fmt.Fprintf(os.Stderr, "Enter pass-phrase: ")
	try1, _ := terminal.ReadPassword(int(f.Fd()))

	if !decrypt {
		fmt.Fprintf(os.Stderr, "\nEnter pass-phrase again: ")
		try2, _ := terminal.ReadPassword(int(f.Fd()))

		if string(try1) != string(try2) {
			log.Fatal("Error: passphrases dont match")
		}
	}

	fmt.Fprintf(os.Stderr, "\n")

	return try1
}

func makeKey(passPhrase []byte, printKey bool) []byte {
	key := make([]byte, 256)

	for i, _ := range key {
		key[i] = byte(i)
	}

	//var x int
	x := 0

	for i, _ := range key {
		x = int(byte(x) + passPhrase[(i%len(passPhrase))] + (key[i] & '\xFF'))
		swap := key[i]
		key[i] = key[x]
		key[x] = swap
	}

	if printKey {
		fmt.Println("key: ", base64.StdEncoding.EncodeToString(key))
	}

	return key
}

func applyEncryption(input []byte, key []byte) []byte {
	output := make([]byte, len(input))

	x := 0
	y := 0

	for i, _ := range input {
		x = (x + 1) % 256
		y = int(key[x] + byte(y)&'\xFF')
		swap := key[x]
		key[x] = key[y]
		key[y] = swap
		r := key[(key[x] + key[y]&'\xFF')]
		output[i] = byte(input[i] ^ r)
	}

	return output
}

func printOrWrite(fname string, suffix string, output []byte) {
	if suffix == "" {
		fmt.Println(string(output))
	} else {
		fnameNew := fname + suffix

		f, err := os.Create(fnameNew)
		if err != nil {
			log.Fatal(err)
		}

		f.Write(output)
	}
}

func main() {
	decrypt, printKey, suffix, fnames := parseArgs()

	passPhrase := readPassPhrase(decrypt)

	key := makeKey(passPhrase, printKey)

	if len(fnames) == 0 {
		fnames = append(fnames, "stdin")
	}

	for _, fname := range fnames {
		input := readFile(fname)

		if decrypt {
			input, _ = base64.StdEncoding.DecodeString(string(input))
		}

		output := applyEncryption(input, key)

		if !decrypt {
			output = []byte(base64.StdEncoding.EncodeToString(output))
		}

		printOrWrite(fname, suffix, output)
	}
}
