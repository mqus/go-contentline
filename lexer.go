package go_contentline

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type pos int

type item struct {
	typ  itemType // The type of this item.
	pos  pos      // The starting position, in bytes, of this item in the input string.
	val  string   // The value of this item.
	line int      // The line number at the start of this item.
}

func (i item) String() string {
	switch {
	case i.typ == itemError:
		return i.val
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError      itemType = iota // error occurred; value is text of error
	itemEquals                     // equals ('=') used for character-assignment
	itemColon                      // :
	itemSemicolon                  // ;
	itemComma                      // ,
	itemParamValue                 // the value of a property parameter, can contain ^^, ^' or ^n
	itemPropValue                  // the value of a property, if the property is of type TEXT, the value can contain \\ , \; , \, , \n or \N
	itemId                         // the Property Name
	itemBegin                      // an indicator for the start of a component
	itemEnd                        // an indicator for the end of a component
	itemCompName                   // the component name
	//itemField      // alphanumeric identifier starting with '.'
	//itemIdentifier // alphanumeric identifier not starting with '.'
	//itemLeftDelim  // left action delimiter
	//itemLeftParen  // '(' inside action
	//itemNumber     // simple number, including imaginary
	//itemPipe       // pipe symbol
	//itemRawString  // raw quoted string (includes quotes)
	//itemRightDelim // right action delimiter
	//itemRightParen // ')' inside action
	//itemSpace      // run of spaces separating arguments
	//itemString     // quoted string (includes quotes)
	//itemText       // plain text
	//itemVariable   // variable starting with '$', such as '$' or  '$1' or '$hello'
	//// Keywords appear after all the rest.
	//itemKeyword  // used only to delimit the keywords
	//itemBlock    // block keyword
	//itemDot      // the cursor, spelled '.'
	//itemDefine   // define keyword
	//itemElse     // else keyword
	//itemEnd      // end keyword
	//itemIf       // if keyword
	//itemNil      // the untyped nil constant, easiest to treat as a keyword
	//itemRange    // range keyword
	//itemTemplate // template keyword
	//itemWith     // with keyword
)
const eof = -1

const (
	wsp     = " 	"
	parName = "-abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	sBEGIN  = "BEGIN"
	sEND    = "END"
)

type lexer struct {
	line  int       // documented for error messages
	input string    // the string being scanned
	pos   pos       // current position in the input
	start pos       // start position of this item
	width pos       // width of last rune read from input
	items chan item // channel of scanned items
}

type stateFn func(*lexer) stateFn

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = pos(w)
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos], l.line}
	// Some items contain text internally. If so, count their newlines.
	l.start = l.pos
}

func (l *lexer) trimQuotesEmit(t itemType) {
	l.items <- item{t, l.start, strings.Trim(l.input[l.start:l.pos], "\""), l.line}
	// Some items contain text internally. If so, count their newlines.
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) acceptUnless(invalid string) bool {
	r := l.next()
	if r == eof || r == 0x7F || (r < 0x20 && r != '\t') || strings.ContainsRune(invalid, r) {
		l.backup()
		return false
	}

	return true
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRunUnless(invalid string) {
	for l.acceptUnless(invalid) {
	}
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...), l.line}
	return nil
}

// nextItem returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) nextItem() item {
	return <-l.items
}

// drain drains the output so the lexing goroutine will exit.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) drain() {
	for range l.items {
	}
}

// lex creates a new scanner for the input string.
func lex(line int, input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan item),
		line:  line,
	}
	go l.run()
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for state := lexPropName; state != nil; {
		state = state(l)
	}
	close(l.items)
}

//state functions

// lexPropName scans until a colon or a semicolon
func lexPropName(l *lexer) stateFn {
	l.acceptRun(parName)
	if l.pos == l.start {
		return l.errorf("expected one or more alphanumerical characters or '-'")
	}
	if strings.ToUpper(l.input[l.start:l.pos]) == sBEGIN {
		l.emit(itemBegin)
		return lexBeforeCompName
	} else if strings.ToUpper(l.input[l.start:l.pos]) == sEND {
		l.emit(itemEnd)
		return lexBeforeCompName
	}
	l.emit(itemId)

	return lexBeforeValue
}

func lexBeforeCompName(l *lexer) stateFn {
	if l.accept(":") {
		l.ignore() //l.emit(itemColon)
		return lexCompName
	}
	return l.errorf("expected ':'")
}

func lexCompName(l *lexer) stateFn {
	if l.peek() == eof {
		return l.errorf("component name can't have length 0")
	}
	l.acceptRun(parName)
	if l.peek() != eof {
		l.ignore()
		return l.errorf("unexpected character, expected eol, alphanumeric or '-'")
	}
	l.emit(itemCompName)
	return nil
}

func lexBeforeValue(l *lexer) stateFn {
	if l.accept(":") {
		l.ignore() //l.emit(itemColon)
		return lexValue
	}
	if l.accept(";") {
		l.ignore() //l.emit(itemSemicolon)
		return lexParamName
	}
	return l.errorf("expected ':' or ';'")
}

func lexParamName(l *lexer) stateFn {
	l.acceptRun(parName)
	if l.pos == l.start {
		return l.errorf("name must not be empty")
	}
	l.emit(itemId)
	if l.accept("=") {
		l.ignore() //l.emit(itemEquals)
		return lexParamValue
	}
	return l.errorf("expected '='")
}

func lexParamValue(l *lexer) stateFn {
	if l.accept("\"") {
		return lexParamQValue
	}
	l.acceptRunUnless("\",;:")
	l.emit(itemParamValue)
	return lexAfterParamValue
}

func lexParamQValue(l *lexer) stateFn {
	l.acceptRunUnless("\"")
	if l.next() != '"' {
		return l.errorf("expected '\"' or other non-control-characters")
	}
	l.trimQuotesEmit(itemParamValue)
	return lexAfterParamValue
}

func lexAfterParamValue(l *lexer) stateFn {
	if l.accept(":") {
		l.ignore() //l.emit(itemColon)
		return lexValue
	}
	if l.accept(";") {
		l.ignore() //l.emit(itemSemicolon)
		return lexParamName
	}
	if l.accept(",") {
		l.ignore() //l.emit(itemComma)
		return lexParamValue
	}
	return l.errorf("expected ',', ':' or ';'")
}

func lexValue(l *lexer) stateFn {
	l.acceptRunUnless("")
	if l.peek() != eof {
		return l.errorf("unexpected character, expected eol")
	}
	l.emit(itemPropValue)
	return nil
}
