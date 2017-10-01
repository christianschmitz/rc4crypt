package main

// standard packages
import (
	"bufio"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path"
)

// It is not worthwhile trying to eliminate the following depencency.
// It is pretty standardized, and gives abstraction of target systems.
import (
	"golang.org/x/crypto/ssh/terminal"
)

const (
	BYTE0   byte = '\x00'
	BYTE256 byte = '\xff'
)

func readStdin() []byte {
	r := bufio.NewReader(os.Stdin)
	bytes, _ := r.ReadBytes(BYTE0)

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
		bytes, _ = r.ReadBytes(BYTE0)
	}

	return bytes
}

func printUsage() {
	// shortcut, to give us cleaner code
	var f = func(str string) {
		fmt.Fprintf(os.Stderr, str)
	}

	f("Usage: rc4crypt [-p | [-h] [-s <suffix>] [<file> ..[<file>]]]\n\n")

	f("Options:\n")
	f(" -p           Print the key generated from the pass-phrase.\n")
	f("              This key can then be used in other crypto tools (e.g. openssl or mcrypt).\n")
	f(" -s <suffix>  Generate output filenames by appending this suffix to input filenames.\n")
	f("              Ignored when reading from stdin.\n")
	f(" -h           Print this message.\n\n")

	f("Details:\n")
	f("  Read from stdin if no files are specified.\n")
	f("  Output is printed to stdout when reading from stdin.\n")
	f("  Output is printed to stdout if no suffix is specified for input files.\n")
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
func parseArgs() (bool, string, []string) {
	i := 0
	args := os.Args[1:]

	printKey := false
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
			case arg == "-p":
				printKey = true

				if n > 1 {
					printUsageAndQuit("Error: the " + arg + " option allows no other options or arguments")
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

	return printKey, suffix, fnames
}

func readPassPhrase(printKey bool) ([]byte, bool) {
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

	msg2 := "Enter pass-phrase again (leave blank to decrypt):"

	if printKey {
		msg2 = "Enter pass-phrase again:"
	}

	fmt.Fprintf(os.Stderr, msg2)
	try2, err2 := terminal.ReadPassword(int(f.Fd()))
	fmt.Fprintf(os.Stderr, "\n")
	if err2 != nil {
		log.Fatal(err2)
	}

	decrypt := false

	if string(try2) == "" {
		if printKey {
			log.Fatal("Error: key generation requires entering pass-phrase twice")
		} else {
			decrypt = true
		}
	} else if string(try1) != string(try2) {
		log.Fatal("Error: pass-phrases dont match")
	}

	return try1, decrypt
}

func makeKey(passPhrase []byte, printKey bool) []byte {
	key := make([]byte, 256)

	for i, _ := range key {
		key[i] = byte(i)
	}

	x := 0

	for i, _ := range key {
		x = int(byte(x) + passPhrase[(i%len(passPhrase))] + (key[i] & BYTE256))
		tmp := key[i]
		key[i] = key[x]
		key[x] = tmp
	}

	if printKey {
		fmt.Printf(base64.StdEncoding.EncodeToString(key))
		os.Exit(0)
	}

	return key
}

func applyEncryption(input []byte, keyOrig []byte) []byte {
	// copy the key so it isn't changed
	key := make([]byte, len(keyOrig))
	for i, v := range keyOrig {
		key[i] = v
	}

	output := make([]byte, len(input))

	x := 0
	y := 0

	for i, _ := range input {
		x = (x + 1) % 256
		y = int(key[x] + byte(y)&BYTE256)
		tmp := key[x]
		key[x] = key[y]
		key[y] = tmp
		r := key[(key[x] + key[y]&BYTE256)]
		output[i] = byte(input[i] ^ r)
	}

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
	printKey, suffix, fnames := parseArgs()

	passPhrase, decrypt := readPassPhrase(printKey)

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
