package main

import (
	"os"
	"fmt"
	"github.com/moshee/mojibake"
	"flag"
	"io"
)

var (
	flag_mode = flag.String("mode", "sjis", "Source encoding (sjis, utf8)")
	flag_out = flag.String("o", "", "Output to file instead of STDOUT")
)

func errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,`Usage: %s [OPTIONS] [infile]
If no input file is given, reads from STDIN.
Options:
`, os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	var err error
	var infile *os.File
	if flag.NArg() == 0 {
		infile = os.Stdin
	} else {
		infile, err = os.Open(flag.Arg(0))
		if err != nil {
			errorf("%s\n", err.Error())
		}
	}

	var outfile *os.File
	if len(*flag_out) == 0 {
		outfile = os.Stdout
	} else {
		outfile, err = os.Create(*flag_out)
		if err != nil {
			errorf("%s\n", err.Error())
		}
	}

	var w io.Writer
	switch *flag_mode {
	case "sjis":
		w = mojibake.SJISDecoder{outfile}
	case "utf8":
		w = mojibake.UTF8Decoder{outfile}
	default:
		fmt.Fprintln(os.Stderr, "Unknown encoding:", *flag_mode)
		flag.Usage()
		os.Exit(1)
	}

	io.Copy(w, infile)
}
