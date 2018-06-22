// Package go-contentline provides an Interface for encoding/decoding files formatted using the same syntactic
// format as vcard and ical files (vcf/ics). The general syntax can be summarized as a structured text file containing
// ContentLines (which describe parametrized properties) and lines marking the start/end of a component.
package go_contentline

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
)

// The maximal Length of a resulting line, any more characters will be folded as described below.
const foldingLength = 75

//Component is the outermost structured part and can include multiple other Components and has Properties.
// There are constraints on most Component types concerning which Properties to include and how often. These will not
// be checked here.
type Component struct {
	//Name is the identifying name for this component, e.g. VCARD or VALARM
	// The component identifiers must be iana-registered tokens or have to be prefixed with 'x-'. This will not be checked!
	// The identifier is case-insensitive and will be converted to uppercase when encoding/parsing.
	Name string

	//Properties contains all included Properties.
	Properties []*Property

	//Comps contains all included Components. For vcf-files, this field should be empty (nil), which will not be checked.
	Comps []*Component
}

//Property is the way to include Values into Components. Properties can also have Parameters.
// E.g. the parameter LANG for DESCRIPTION describes the language in which the description is written.
type Property struct {
	//Name is the identifying name for this property, e.g. DESCRIPTION or ROLE
	// The property identifiers must be iana-registered tokens or have to be prefixed with 'x-'. This will not be checked!
	// The identifier is case-insensitive and will be converted to uppercase when encoding/parsing.
	Name string

	//Value is the value for this property, depending on the Name it can have one of multiple types which
	// includes varying restrictions on the format. The Value will be encoded/parsed as-is, meaning without any
	// (un-)escaping of newline characters and so on. Therefore this string must not contain any newline
	// characters (0x0a and 0x0d), as well as any control characters besides HTAB(0x09)
	Value string

	//Parameters contains the property parameters. For details, see below.
	Parameters

	//field for remembering the original form before parsing, see Property.OriginalLine()
	olds string
}

func NewProperty(name, value string, p Parameters) (out *Property, err error) {
	r := ValidID(name)
	if r == nil {
		return NewPropertyUnchecked(name, value, p), nil
	} else {
		return nil, errors.Errorf("Contains at least one illegal character:'%v'", r)
	}
}

func NewPropertyUnchecked(name, value string, p Parameters) *Property {
	return &Property{name, value, p, ""}
}

//Parameters is a type to represent property parameters as described
// by RFC6868; RFC5545, Section 3.2 and RFC6350, Section 5. The parameter identifiers must
// be iana-registered tokens or have to be prefixed with 'x-'. This will not be checked!
// The identifiers are case-insensitive and will be converted to uppercase when encoding/parsing.
// The parameter values can include any utf8-codepoint, as long as they are not control
// characters (ASCII 0x00 - 0x08,0x0b,0x0c and 0x0e-0x1f), BUT depending on the parameter name a standard could define
// more constraints (e.g. only a defined set of values for VALUE)
type Parameters map[string][]string

//OriginalLine returns the unfolded line from the input, before it was parsed.
// That can be useful for error messages in further conversion into calendar/contact objects.
// This method will return an empty string if this Property was not parsed, but created
func (p *Property) OriginalLine() string {
	return p.olds
}

//Encode encodes the component as described in RFC5545, Section 3.4 and 3.6ff or also RFC6350, Section 6.1.1/6.1.2,
// including encoding all Properties and writes it to the Writer interface. This writer must be closed by the calling function
// and is left open for more objects.
func (c *Component) Encode(w io.Writer) {
	fmt.Fprintf(w, "%s:%s\r\n", sBEGIN, strings.ToUpper(c.Name))
	for _, p := range c.Properties {
		p.Encode(w)
	}
	for _, c := range c.Comps {
		c.Encode(w)
	}

	fmt.Fprintf(w, "%s:%s\r\n", sEND, strings.ToUpper(c.Name))
}

//Encode encodes the property to a contentline as described in RFC5545, Section 3.1 or also RFC6350, Section 3.3,
// folds it (if neccessary) and writes it to the Writer interface. This writer must be closed by the calling function
// and is left open for more objects.
func (p *Property) Encode(w io.Writer) {
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

//writeFolded folds the ContentLine (s) as described in RFC5545, Section 3.1 or also RFC6350, Section 3.2
// and then writes it to the given Writer interface.
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
