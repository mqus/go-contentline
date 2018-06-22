package go_contentline

import (
	"strings"

	"unicode/utf8"

	"io"

	"github.com/pkg/errors"
)

//ValidID checks if a string is a valid identifier (iana-token or x-token).
// If not, returns the conflicting rune.
func ValidID(in string) *rune {
	for _, r := range []rune(in) {
		if !strings.ContainsRune(parName, r) {
			return &r
		}
	}
	return nil
}

//contentRadius defines the amount of characters which will be displayed around the token causing the error
const contentRadius = 20

//errorf is a helper function generating pretty error messages for errors thrown py the parser.
func errorf(line string, i *item, msg string) error {

	prefix := ""
	suffix := ""

	pos1 := i.pos
	pos2 := i.pos + pos(len(i.val))
	if msg == "" {
		msg = i.val
		pos2 = pos1 + 1
	}

	if pos1 > contentRadius {
		prefix = "..." + line[pos1-contentRadius:pos1]
	} else {
		prefix = line[0:pos1]
	}

	if len(line) > contentRadius+int(pos2) {
		suffix = line[pos1:pos2+contentRadius] + "..."
	} else if int(pos1) < len(line) {
		suffix = line[pos1:]
	}

	if len(suffix) == 0 {
		return errors.Errorf("%s: \t%s<HERE>\n", msg, prefix)
	}
	if len(suffix) == 1 {
		return errors.Errorf("%s: \t%s >%s<\n", msg, prefix, suffix[:pos2-pos1])
	}
	return errors.Errorf("%s: \t%s >%s< %s\n", msg, prefix, suffix[:pos2-pos1], suffix[pos2-pos1:])
}

//split the input string in parts which are at most maxlen bytes long, while preserving utf8-runes
func split(in string, maxlen int) (out []string) {
	if len(in) <= maxlen {
		return []string{in}
	}
	inr := []rune(in)
	out = nil
	prev := 0
	sum := 0
	for _, r := range inr {
		rl := utf8.RuneLen(r)
		if sum+rl-prev > maxlen {
			//decrease maxlen for the space which will be added in writeFolded
			if prev == 0 {
				maxlen--
			}
			out = append(out, in[prev:sum])
			prev = sum
		}
		sum = sum + rl
	}
	out = append(out, in[prev:])
	return
}

//needed helper for the example in lex_test.go
type filterwrite struct {
	ignore byte
	inner  io.Writer
}

func (f *filterwrite) Write(p []byte) (n int, err error) {
	for i, b := range p {
		if b != f.ignore {
			x, e := f.inner.Write([]byte{b})
			if e != nil {
				return i + x, e
			}
		}
	}
	return len(p), nil
}

func filter(w io.Writer, ignore byte) io.Writer {
	return &filterwrite{ignore, w}
}
