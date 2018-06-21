package go_contentline

import (
	"bufio"
	"bytes"

	"io"

	"strings"

	"github.com/pkg/errors"
)

type Parser struct {
	r    *bufio.Reader
	line int //not useful
	l    *lexer
}

func (p *Parser) readUnfoldedLine() (string, error) {
	buf, e := p.r.ReadBytes('\r')
	if e != nil {
		return "", e
	}

	b, err := p.r.ReadByte()
	if err != nil {
		return "", err
	}
	if b != '\n' {
		return "", errors.New("Expected CRLF:" + string(buf))
	}
	b1, err2 := p.r.Peek(1)

	if err2 != nil {
		return string(buf[:len(buf)-1]), err2
	}
	if bytes.Equal(b1, []byte(" ")) {
		p.r.ReadByte()
		s, e := p.readUnfoldedLine()
		if s == "" {
			return "", e
		}
		return string(buf[:len(buf)-1]) + s, e
	}
	return string(buf[:len(buf)-1]), nil
}

func InitParser(reader io.Reader) Parser {
	return Parser{bufio.NewReader(reader), 0, nil}
}

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
		e := errorf(p.l, &i, "")
		p.l = nil
		return nil, e

	case itemCompName, itemPropValue: //the last items of a line
		p.l = nil
	case itemId: // make it easier for string matching
		i.val = strings.ToUpper(i.val)
	case itemParamValue: // remove escape strings (^^,^n,^N,^')
		i.val = UnescapeParamVal(i.val)
	}
	return &i, nil

}

func (p *Parser) ParseComponent() (component *Component, err error) {
	var i *item
	i, err = p.getNextItem()
	if err != nil {
		return nil, err
	}
	if i.typ != itemBegin {
		return nil, errorf(p.l, i, "Expected 'BEGIN'")
	}

	return p.parseComponent()
}

//BEGIN was already read
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
	for ; i.typ != itemEnd; i, e = p.getNextItem() {
		if e != nil {
			return nil, e
		}
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
	namei, err := p.getNextItem()
	if err != nil {
		return nil, err
	}
	if namei.val != out.Name {
		return nil, errorf(p.l, namei, "expected "+out.Name)
	}
	return out, nil
}

func (p *Parser) parseProperty(name string) (*Property, error) {
	out := &Property{
		Name:       name,
		Parameters: make(map[string][]string),
		olds:       p.l.input,
	}

	currentParam := ""
	i, e := p.getNextItem()
	for ; i.typ != itemPropValue; i, e = p.getNextItem() {
		if e != nil {
			return nil, e
		}
		switch i.typ {
		case itemId:
			currentParam = i.val
		case itemParamValue:
			out.Parameters[currentParam] = append(out.Parameters[currentParam], i.val)
		}
	}
	out.Value = i.val
	return out, nil
}

func errorf(l *lexer, i *item, msg string) error {
	prefix := ""
	suffix := ""

	pos1 := i.pos
	pos2 := i.pos + Pos(len(i.val))
	if msg == "" {
		msg = i.val
		pos2 = pos1 + 1
	}

	if pos1 > 10 {
		prefix = "..." + l.input[pos1-10:pos1]
	} else {
		prefix = l.input[0:pos1]
	}

	if len(l.input) > 10+int(pos2) {
		suffix = l.input[pos1:pos2+10] + "..."
	} else {
		suffix = l.input[pos2:]
	}

	return errors.Errorf("%s: \t%s >%s< %s\n", msg, prefix, suffix[:pos2-pos1], suffix[pos2-pos1:])
}
