package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/moshee/mojibake"
	"io"
	"os"
	"strings"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `Available encoding options for -encs:
    utf-8, utf8, cp473: CP473 (assume UTF-8 input misinterpreted as ASCII)
    sjis, shift-jis, cp932: CP932 (assume Shift-JIS input misinterpreted as UTF-8)
    cjk, cp936: CP936 (assume CJK input misinterpreted as UTF-8)
`)
	}
}

var (
	flag_encs   = flag.String("encs", "utf-8", "Comma-separated encoding path")
	flag_args   = flag.Bool("args", false, "Decode arguments instead of from STDIN")
	flag_rename = flag.Bool("rename", false, "Like args but rename the named files to the decoded values")
	flag_really = flag.Bool("really", false, "When -rename is given, actually do the renaming instead of just showing what will happen")
	flag_force  = flag.Bool("f", false, "Skip errors")
	//flag_recurse = flag.Bool("r", false, "When -rename is given, perform renaming recursively through the given directories (THIS IS REALLY DANGEROUS)")
)

var enc_map = map[string]mojibake.Encoding{
	"utf-8":     mojibake.CP473,
	"utf8":      mojibake.CP473,
	"cp473":     mojibake.CP473,
	"sjis":      mojibake.CP932,
	"shift-jis": mojibake.CP932,
	"cp932":     mojibake.CP932,
	"cjk":       mojibake.CP936,
	"cp936":     mojibake.CP936,
}

func main() {
	flag.Parse()

	if len(*flag_encs) == 0 {
		fmt.Fprintln(os.Stderr, "unmoji: no encoding path given")
		os.Exit(1)
	}
	enc_list := strings.Split(*flag_encs, ",")
	encs := make([]mojibake.Encoding, 0, len(enc_list))
	for _, enc_name := range enc_list {
		if enc, ok := enc_map[enc_name]; ok {
			encs = append(encs, enc)
		} else {
			fmt.Fprintf(os.Stderr, "unmoji: unknown encoding: %s\n", enc_name)
			os.Exit(1)
		}
	}

	if *flag_args || *flag_rename {
		os.Exit(decode_args(encs))
	} else {
		os.Exit(decode_stdin(encs))
	}
}

func decode_args(encs []mojibake.Encoding) int {
	if flag.NArg() == 0 {
		return 0
	}

	buf := new(bytes.Buffer)
	dec, err := mojibake.NewDecoder(buf, encs...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unmoji: %v\n", err)
		return 1
	}

	decoded := make([]string, 0, flag.NArg())
	args := flag.Args()

	// need this in case any of them error and -f is given
	// so we can selectively include input filenames
	input := make([]string, 0, len(args))

	for _, garbled := range args {
		r := strings.NewReader(garbled)
		io.Copy(dec, r)

		_, err := dec.Flush()
		if err != nil {
			if *flag_force {
				continue
			}
			fmt.Fprintf(os.Stderr, "unmoji: %v\n", err)
			dec.Close()
			return 1
		}

		p := buf.Next(buf.Len())
		input = append(input, garbled)
		decoded = append(decoded, string(p))
	}

	dec.Close()

	if !*flag_rename {
		for _, s := range decoded {
			fmt.Println(s)
		}

		return 0
	}

	if !*flag_really {
		for i, filename := range input {
			fmt.Printf("\"%s\" → \"%s\"\n", filename, decoded[i])
		}

		return 0
	}

	for i, filename := range input {
		if filename == decoded[i] && !*flag_force {
			fmt.Fprintf(os.Stderr, "unmoji: rename \"%s\": source and destination are the same\n", filename)
			return 1
		}

		err := os.Rename(filename, decoded[i])
		if err != nil && !*flag_force {
			fmt.Fprintf(os.Stderr, "unmoji: %v\n", err)
			return 1
		}
	}

	return 0
}

func decode_stdin(encs []mojibake.Encoding) int {
	decoder, err := mojibake.NewDecoder(os.Stdout, encs...)
	if err != nil && !*flag_force {
		fmt.Fprintf(os.Stderr, "unmoji: %v\n", err)
		return 1
	}

	io.Copy(decoder, os.Stdin)

	err = decoder.Close()
	if err != nil && !*flag_force {
		fmt.Fprintf(os.Stderr, "unmoji: %v\n", err)
		return 1
	}

	return 0
}
