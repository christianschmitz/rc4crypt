// copyright: Christian Schmitz
// contact: christian.schmitz@telenet.be

package main

// standard packages
import (
	"bufio"
	"crypto/rc4"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path"
)

import (
	// it is not worthwhile trying to eliminate the following depencency
	// (it is pretty standardized, and gives abstraction of target systems)
	"golang.org/x/crypto/ssh/terminal"
)

const STOPBYTE byte = '\x00'

func readStdin() []byte {
	r := bufio.NewReader(os.Stdin)
	bytes, _ := r.ReadBytes(STOPBYTE)

	return bytes
}

func readFile(fname string) []byte {
	var bytes []byte

	if fname == "stdin" {
		bytes = readStdin()
	} else {
		f, err := os.Open(fname)
		defer f.Close()
		if err != nil {
			log.Fatal(err)
		}

		r := bufio.NewReader(f)
		bytes, _ = r.ReadBytes(STOPBYTE)
	}

	return bytes
}

func printUsage() {
	// shortcut, to give us cleaner code
	var f = func(str string) {
		fmt.Fprintf(os.Stderr, str)
	}

	f("Usage: rc4crypt [-h] [-s <suffix>] [<file> [<file> ..]]\n\n")

	f("Options:\n")
	f(" -s <suffix>  Generate output filename(s) by appending this suffix to input filename(s).\n")
	f("              Ignored when reading from stdin.\n")
	f(" -h           Print this message.\n\n")

	f("Details:\n")
	f("  Read from stdin if no files are specified.\n")
	f("  Output is printed to stdout when reading from stdin.\n")
	f("  Output is printed to stdout if no suffix is specified for input file(s).\n")
	f("  Pass-phrase is read from terminal.\n")
	f("  Decrypt by entering blank pass-phrase at second prompt.\n")
	f("  Uses rc4 (a.k.a. arcfour) algorithm and base64 encoding.\n")
	f("  Encrypts/decrypts plain text into plain text.\n\n")
}

func printUsageAndQuit(msg string) {
	exitCode := 0
	if msg != "" { // msg is assumed to represent an error
		fmt.Fprintf(os.Stderr, msg+"\n")
		exitCode = 1
	}

	printUsage()

	os.Exit(exitCode)
}

// using the flag package doesnt allow putting options after other arguments
// this can be annoying
func parseArgs() (string, []string) {
	i := 0
	args := os.Args[1:]

	suffix := ""
	fnames := make([]string, 0)

	n := len(args)
	// loop the arguments
	for i < n {
		arg := args[i]

		if arg[0] == '-' { // options
			switch {
			case arg == "-s":
				if i < n-1 {
					suffix = args[i+1]
					i = i + 1
				} else {
					printUsageAndQuit("Error: the " + arg + " option requires an argument")
				}
			case arg == "-h":
				printUsageAndQuit("")
			default:
				printUsageAndQuit("Error: option " + arg + " not recognized")
			}
		} else { // positional arguments (i.e. filenames)
			// check for file existence as soon as possible
			f, err := os.Open(arg)
			if err != nil {
				printUsageAndQuit("Error: " + arg + " is not a file")
			}
			f.Close()

			fnames = append(fnames, arg)
		}

		i = i + 1
	}

	return suffix, fnames
}

func readPassPhrase() ([]byte, bool) {
	f, err := os.Open("/dev/tty")
	defer f.Close()
	if err != nil {
		f = os.Stdin
	}

	fmt.Fprintf(os.Stderr, "Enter pass-phrase: ")
	try1, err1 := terminal.ReadPassword(int(f.Fd()))
	fmt.Fprintf(os.Stderr, "\n")
	if err1 != nil {
		log.Fatal(err1)
	}
	if string(try1) == "" {
		log.Fatal("Error: pass-phrase must be at least 1 character long")
	}

	fmt.Fprintf(os.Stderr, "Enter pass-phrase again (leave blank to decrypt):")
	try2, err2 := terminal.ReadPassword(int(f.Fd()))
	fmt.Fprintf(os.Stderr, "\n")
	if err2 != nil {
		log.Fatal(err2)
	}

	decrypt := false

	if string(try2) == "" {
		decrypt = true
	} else if string(try1) != string(try2) {
		log.Fatal("Error: pass-phrases dont match")
	}

	return try1, decrypt
}

func applyEncryption(input []byte, passPhrase []byte) []byte {
	cipher, err := rc4.NewCipher(passPhrase)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: pass-phrase longer than 256 bytes, truncating")
	}

	output := make([]byte, len(input))

	cipher.XORKeyStream(output, input)

	return output
}

func printOrWrite(fname string, suffix string, output []byte) {
	if suffix == "" || fname == "stdin" {
		fmt.Println(string(output))
	} else {
		fnameNew := path.Base(fname + suffix)

		f, err := os.Create(fnameNew)
		if err != nil {
			log.Fatal(err)
		}

		f.Write(output)
		f.Close()
	}
}

func main() {
	suffix, fnames := parseArgs()

	passPhrase, decrypt := readPassPhrase()

	if len(fnames) == 0 {
		fnames = append(fnames, "stdin")
	}

	for _, fname := range fnames {
		input := readFile(fname)

		if decrypt {
			input, _ = base64.StdEncoding.DecodeString(string(input))
		}

		output := applyEncryption(input, passPhrase)

		if !decrypt {
			output = []byte(base64.StdEncoding.EncodeToString(output))
		}

		printOrWrite(fname, suffix, output)
	}
}
