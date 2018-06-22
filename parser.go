package go_contentline

import (
	"bufio"
	"bytes"

	"io"

	"strings"

	"github.com/pkg/errors"
)

//Parser contains fields describing the state of the parser.
type Parser struct {
	r    *bufio.Reader
	line int //not useful
	l    *lexer
}

//InitParser initializes the parser by creating a buffered Reader.
func InitParser(reader io.Reader) *Parser {
	return &Parser{bufio.NewReader(reader), 0, nil}
}

//ParseNextObject parses the next Component and returns it. If the Parser encounters an EOF prematurely,
// it returns 'nil, io.EOF'. For all other errors, a wrapped error is returned.
func (p *Parser) ParseNextObject() (component *Component, err error) {
	c, e := p.parseObject()
	switch e {
	case nil:
		return c, nil
	case io.EOF:
		return nil, io.EOF
	default:
		return nil, errors.Wrap(e, "error while parsing component(s)")
	}
}

//parseObject parses the next Object from the stream, expecting an itemBegin token. This function is wrapped by
// ParseNextObject for better error messages.
func (p *Parser) parseObject() (component *Component, err error) {
	var i *item
	//checks if the first thing to read is the start of a component
	i, err = p.getNextItem()
	if err != nil {
		return nil, err
	}
	if i.typ != itemBegin {
		return nil, errorf(p.l.input, i, "Expected '"+sBEGIN+"'")
	}
	//if true, start recursively parsing components and properties
	return p.parseComponent()
}

//parseComponent parses the Component for which itemBegin was already read.
func (p *Parser) parseComponent() (*Component, error) {
	var i *item
	var err error
	i, err = p.getNextItem()
	if err != nil {
		return nil, err
	}

	out := &Component{
		Name: i.val,
	}

	i, e := p.getNextItem()
	for ; e == nil && i.typ != itemEnd; i, e = p.getNextItem() {
		switch i.typ {
		case itemId:
			p, e := p.parseProperty(i.val)
			if e != nil {
				return nil, e
			}

			out.Properties = append(out.Properties, p)

		case itemBegin:
			c, e := p.parseComponent()
			if e != nil {
				return nil, e
			}

			out.Comps = append(out.Comps, c)
		}
	}
	if e != nil {
		return nil, e
	}
	line := p.l.input
	namei, err := p.getNextItem()
	if err != nil {
		return nil, err
	}
	if namei.val != out.Name {
		return nil, errorf(line, namei, "expected "+out.Name)
	}
	return out, nil
}

//parseProperty parses the next Property while already having parsed the Property name.
func (p *Parser) parseProperty(name string) (*Property, error) {
	out := &Property{
		Name:       name,
		Parameters: make(map[string][]string),
		olds:       p.l.input,
	}

	currentParam := ""
	i, e := p.getNextItem()
	for ; e == nil && i.typ != itemPropValue; i, e = p.getNextItem() {
		switch i.typ {
		case itemId:
			currentParam = i.val
		case itemParamValue:
			out.Parameters[currentParam] = append(out.Parameters[currentParam], i.val)
		}
	}
	if e != nil {
		return nil, e
	}
	out.Value = i.val
	return out, nil
}

//getNextItem returns the next lexer item, feeding (unfolded) lines into the lexer if neccessary.
// It also converts identifiers (itemCompName, itemID) into upper case, errors encountered by the
// lexer into 'error' values and property parameter values into their original value (without escaped characters).
func (p *Parser) getNextItem() (*item, error) {
	if p.l == nil {
		line, err := p.readUnfoldedLine()
		if line == "" {
			return nil, err
		}
		p.l = lex(p.line, line)
	}
	i := p.l.nextItem()
	switch i.typ {
	case itemError:
		e := errorf(p.l.input, &i, "")
		p.l = nil
		return nil, e
	case itemCompName:
		i.val = strings.ToUpper(i.val)
		fallthrough
	case itemPropValue: //the last items of a line
		p.l = nil
	case itemId: // make it easier for string matching
		i.val = strings.ToUpper(i.val)
	case itemParamValue: // remove escape strings (^^,^n,^N,^')
		i.val = UnescapeParamVal(i.val)
	}
	return &i, nil

}

//readUnfoldedLine reads lines directly from the reader and unfolds them if neccessary.
func (p *Parser) readUnfoldedLine() (string, error) {
	buf, e := p.r.ReadBytes('\n')
	if e != nil {
		return "", e
	}

	if buf[len(buf)-2] != '\r' {
		return "", errors.Errorf("Expected CRLF:%s, >%v<", buf, buf[len(buf)-2:])
	}
	b1, err2 := p.r.Peek(1)

	if err2 != nil {
		return string(buf[:len(buf)-2]), err2
	}
	if bytes.Equal(b1, []byte(" ")) || bytes.Equal(b1, []byte("\t")) {
		p.r.ReadByte()
		s, e := p.readUnfoldedLine()
		if s == "" {
			return "", e
		}
		return string(buf[:len(buf)-2]) + s, e
	}
	return string(buf[:len(buf)-2]), nil
}

//contentRadius defines the amount of characters which will be displayed around the token causing the error
const contentRadius = 20

//errorf is a helper function generating pretty error messages for errors thrown py the parser.
func errorf(line string, i *item, msg string) error {

	prefix := ""
	suffix := ""

	pos1 := i.pos
	pos2 := i.pos + Pos(len(i.val))
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
