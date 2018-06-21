package go_contentline

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

const foldingLength = 75

type Component struct {
	Name       string
	Comps      []*Component
	Properties []*Property
}

type Property struct {
	Name  string
	Value string
	Parameters
	olds string
}
type Parameters map[string][]string

func (p *Property) BeforeParsing() string {
	//old must be the read string representation of this Property ("ContentLine")
	return p.olds
}

func (c *Component) EncodeICal(w io.Writer) {
	fmt.Fprintf(w, "%s:%s\r\n", sBEGIN, strings.ToUpper(c.Name))
	for _, p := range c.Properties {
		p.EncodeICal(w)
	}
	for _, c := range c.Comps {
		c.EncodeICal(w)
	}

	fmt.Fprintf(w, "%s:%s\r\n", sEND, strings.ToUpper(c.Name))
}

func (p *Property) EncodeICal(w io.Writer) {
	out := strings.ToUpper(p.Name)
	//log.Printf("NoPARAM: %v\n", p.Parameters)
	for k, vals := range p.Parameters {
		//log.Println("INPARAM")
		out = out + ";" + strings.ToUpper(k) + "="
		for i, v := range vals {
			if i > 0 {
				out = out + ","
			}
			val := EscapeParamVal(v)
			if strings.ContainsAny(val, ",;:") {
				val = "\"" + val + "\""
			}
			out = out + val
		}
	}
	out = out + ":" + p.Value
	writeFolded(w, out)
}

func writeFolded(w io.Writer, s string) {
	parts := split(s, foldingLength)
	for i, part := range parts {
		if i > 0 {
			fmt.Fprint(w, " ")
		}
		fmt.Fprint(w, part)
		fmt.Fprint(w, "\r\n")
	}
}

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
			//decrease maxlen for the space which will be added
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
