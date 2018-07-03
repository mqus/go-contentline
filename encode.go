// Package go-contentline provides an Interface for encoding/decoding files formatted using the same syntactic
// format as vcard and ical files (vcf/ics). The general syntax can be summarized as a structured text file containing
// ContentLines (which describe parametrized properties) and lines marking the start/end of a component.
package go_contentline

import (
	"fmt"
	"io"
	"strings"
)

// The maximal Length of a resulting line, any more characters will be folded as described below.
const foldingLength = 75

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
