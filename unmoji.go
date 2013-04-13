package main

import (
	"flag"
	"fmt"
	"github.com/moshee/mojibake"
	"io"
	"os"
	"unicode/utf8"
)

var (
	flag_mode = flag.String("mode", "", "Source encoding ('' (convert), 'sjis', 'utf8', 'cjk')")
	flag_out  = flag.String("o", "", "Output to file instead of STDOUT")
)

func errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: %s [OPTIONS] [infile]
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
	case "":
		w = mojibake.SJISDecoder{outfile}
	case "utf8":
		w = mojibake.UTF8Decoder{outfile}
	case "cjk":
		decode(infile, outfile, enc_cp936)
		return
	case "sjis":
		decode(infile, outfile, enc_cp932)
		return

	default:
		fmt.Fprintln(os.Stderr, "Unknown encoding:", *flag_mode)
		flag.Usage()
		os.Exit(1)
	}

	io.Copy(w, infile)
}

func decode(infile io.Reader, outfile *os.File, enc Encoding) {
	var (
		buf = make([]rune, 4096)
		d   = new_decoder(infile, enc)
		ch  rune
		err error
	)
	for i := 0; ; i++ {
		ch, _, err = d.ReadRune()
		if err != nil {
			outfile.WriteString(string(buf[:i]))
			break
		}
		if i >= len(buf) {
			outfile.WriteString(string(buf))
			i = 0
		}
		buf[i] = ch
	}
}

type Encoding int

const (
	enc_cp932 Encoding = iota
	enc_cp936
)

type Decoder struct {
	r     io.Reader
	buf   []byte
	ptr   int
	err   error
	table []rune
}

func new_decoder(r io.Reader, enc Encoding) (s *Decoder) {
	s = new(Decoder)
	s.r = r
	switch enc {
	case enc_cp932:
		s.table = cp932[:]
	case enc_cp936:
		s.table = cp936[:]
	default:
		return nil
	}
	s.buf = make([]byte, 4096)
	s.fill()
	return
}

func (self *Decoder) fill() {
	n, err := self.r.Read(self.buf)
	self.buf = self.buf[:n]
	self.err = err
	self.ptr = 0
}

// Transparent Read function. Does no character transformation. Use ReadRune for actual decoding.
func (self *Decoder) Read(p []byte) (n int, err error) {
	n, err = self.r.Read(p)
	self.ptr += n
	return
}

func (self *Decoder) ReadByte() (byte, error) {
	if self.ptr >= len(self.buf) {
		if self.err == io.EOF {
			return 0, self.err
		}
		self.fill()
		if self.err != nil {
			return 0, self.err
		}
	}
	b := self.buf[self.ptr]
	self.ptr++
	return b, nil
}

func (self *Decoder) ReadRune() (rune, int, error) {
	c1, err := self.ReadByte()
	if err != nil {
		return 0, 0, err
	}
	// converting begins here
	if c1 < utf8.RuneSelf {
		return rune(c1), 1, nil
	}
	// read up to 2 bytes (16 bits) for the 16-bit Shift-JIS character
	c2, err := self.ReadByte()
	if err != nil {
		return utf8.RuneError, 1, err
	}
	// shift-jis is easy since every rune is either one or two bytes
	ch := self.table[rune(c1)<<8|rune(c2)]

	return ch, utf8.RuneLen(ch), nil
}
