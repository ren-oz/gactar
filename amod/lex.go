package amod

// Mostly based on Rob Pike's talk:
// 	https://www.youtube.com/watch?v=HxaD_trXwRE
// Not sure I implemented precisely what he's advocating since I ended
// up with a central switch anyways. I don't see how it can be avoided.

import (
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/alecthomas/participle/v2/lexer"
)

type lexer_def struct {
	lexer.Definition
}

// LexerDefinition provides the interface for the participle parser
var LexerDefinition lexer.Definition = lexer_def{}

type lexemeType int
type lexeme struct {
	typ   lexemeType
	value string
	line  int // line number this lexeme is on
	pos   int // position within the line
}

// lexer_amod tracks our lexing and provides a channel to emit lexemes
type lexer_amod struct {
	name           string // used only for error reports
	input          string // the string being scanned.
	line           int    // the line number
	lastNewlinePos int
	start          int             // start position of this lexeme
	pos            int             // current position in the input
	width          int             // width of last rune read from input
	lexemes        chan lexeme     // channel of scanned lexemes
	keywords       map[string]bool // used to lookup identifier to see if they are keywords
	inDoBlock      bool            // state: a "do" block is lexed as a series of strings
}

// stateFn is used to move through the lexing states
type stateFn func(*lexer_amod) stateFn

const (
	lexemeError lexemeType = iota

	lexemeSpace
	LexemeEOF

	lexemeComment
	lexemeIdentifier
	lexemeKeyword
	lexemeNumber
	lexemeString
	lexemeChar

	lexemeSectionModel
	lexemeSectionConfig
	lexemeSectionInit
	lexemeSectionProductions

	lexemeCodeBegin // marks beginning of code in "do" block
	lexemeDoCode
	lexemeCodeEnd // marks end of code in "do" block
)

const (
	eof = -1

	commentDelim = "//"
	codeBegin    = "#<"
	codeEnd      = ">#"

	sectionModel       = "==model=="
	sectionConfig      = "==config=="
	sectionInit        = "==init=="
	sectionProductions = "==productions=="
)

var keywords []string = []string{
	"actr",
	"arg",
	"buffers",
	"description",
	"do",
	"examples",
	"field",
	"from",
	"match",
	"memories",
	"name",
	"of",
	"print",
	"recall",
	"set",
	"text_outputs",
	"to",
	"write",
}

// Symbols provides a mapping from participle strings to our lexemes
func (lexer_def) Symbols() map[string]lexer.TokenType {
	return map[string]lexer.TokenType{
		"Comment":    lexer.TokenType(lexemeComment),
		"Whitespace": lexer.TokenType(lexemeSpace),
		"Keyword":    lexer.TokenType(lexemeKeyword),
		"Ident":      lexer.TokenType(lexemeIdentifier),
		"Number":     lexer.TokenType(lexemeNumber),
		"String":     lexer.TokenType(lexemeString),
		"DoCode":     lexer.TokenType(lexemeDoCode),
	}
}

// Lex is called by the participle parser to lex a reader
func (lexer_def) Lex(filename string, r io.Reader) (lexer.Lexer, error) {
	s := &strings.Builder{}
	_, err := io.Copy(s, r)
	if err != nil {
		return nil, err
	}

	data := s.String()
	cleanData(&data)

	l := &lexer_amod{
		name:           filename,
		input:          data,
		line:           1,
		lastNewlinePos: 0,
		lexemes:        make(chan lexeme),
		keywords:       make(map[string]bool),
		inDoBlock:      false,
	}

	for _, v := range keywords {
		l.keywords[v] = true
	}

	go l.run()

	return l, nil
}

// Next is used by participle to get the next token
func (l *lexer_amod) Next() (tok lexer.Token, err error) {
	next := <-l.lexemes

	pos := lexer.Position{
		Filename: l.name,
		Offset:   l.pos,
		Line:     next.line,
		Column:   next.pos,
	}

	if next.typ == LexemeEOF {
		return lexer.EOFToken(pos), nil
	}

	tok = lexer.Token{
		Type:  lexer.TokenType(next.typ),
		Value: next.value,
		Pos:   pos,
	}

	if next.typ == lexemeError {
		err = fmt.Errorf("ERROR on line %d at position %d: %s", next.line, next.pos, next.value)
		return
	}

	if debugging {
		fmt.Printf("TOK: %+v (%d)\n", tok, tok.Type)
	}
	return
}

func (l *lexer_amod) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}

	r, width := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = width
	l.pos += l.width

	return r
}

func (l *lexer_amod) lookupKeyword(id string) bool {
	v, ok := l.keywords[id]
	return v && ok
}

// skip over the pending input before this point
func (l *lexer_amod) ignore() {
	l.start = l.pos
}

// step back one rune
func (l *lexer_amod) backup() {
	l.pos -= l.width
}

// look at the next rune in the input, but don't eat it
func (l *lexer_amod) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// check if next rune is "r"
func (l *lexer_amod) nextIs(r rune) bool {
	return l.peek() == r
}

// accept any character in the string
func (l *lexer_amod) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}

	l.backup()

	return false
}

// accept a run of any characters in the string
func (l *lexer_amod) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}

	l.backup()
}

// pass an item back to the client via the channel
func (l *lexer_amod) emit(t lexemeType) {
	l.lexemes <- lexeme{
		typ:   t,
		value: l.input[l.start:l.pos],
		line:  l.line,
		pos:   l.pos - l.lastNewlinePos,
	}

	l.start = l.pos
}

// declare and error and let the client know where we are in the input
func (l *lexer_amod) errorf(format string, args ...interface{}) stateFn {
	l.lexemes <- lexeme{
		lexemeError,
		fmt.Sprintf(format, args...),
		l.line,
		l.pos - l.lastNewlinePos,
	}

	return nil
}

func (l *lexer_amod) run() {
	for state := lexStart; state != nil; {
		// name := runtime.FuncForPC(reflect.ValueOf(state).Pointer()).Name()
		// fmt.Printf("%s\n", name)

		state = state(l)
	}

	close(l.lexemes)
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n'
}

// newlines have been normalized, so just check the one
func isNewline(r rune) bool {
	return r == '\n'
}

func isAlphaNumeric(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func isDigit(r rune) bool {
	return ('0' <= r && r <= '9')
}

func lexStart(l *lexer_amod) stateFn {
	switch r := l.next(); {
	case isSpace(r):
		if isNewline(r) {
			l.lastNewlinePos = l.pos + 1
			l.line++
		}
		return lexSpace

	case isDigit(r):
		l.backup()
		return lexNumber

	case (r == '+') || (r == '-'):
		if isDigit(l.peek()) {
			l.backup()
			return lexNumber
		}
		l.emit(lexemeChar)

	case isAlphaNumeric(r):
		return lexIdentifier

	case r == '=':
		if l.nextIs('=') {
			l.backup()
			return lexSection
		}
		l.emit(lexemeChar)

	case r == '/':
		if l.nextIs('/') {
			l.backup()
			return lexComment
		}
		l.backup()

	case r == '"' || r == '\'':
		l.backup()
		return lexQuotedString

	case r == '#':
		if l.nextIs('<') {
			l.next()
			l.inDoBlock = true
			l.emit(lexemeCodeBegin)

			return lexDoBlock
		}
		l.emit(lexemeChar)

	case r <= unicode.MaxASCII && unicode.IsPrint(r):
		l.emit(lexemeChar)

	case r == eof:
		l.emit(LexemeEOF)
		return nil
	}

	return lexStart
}

// consume 0 or more spaces
func eatSpace(l *lexer_amod) {
	for {
		r := l.next()

		if !isSpace(r) {
			l.backup()
			break
		}

		if isNewline(r) {
			l.lastNewlinePos = l.pos + 1
			l.line++
		}
	}
	l.ignore()
}

func lexSpace(l *lexer_amod) stateFn {
	eatSpace(l)
	return lexStart
}

func lexComment(l *lexer_amod) stateFn {
	l.pos += len(commentDelim)
	i := strings.Index(l.input[l.pos:], "\n")
	l.pos += i

	l.emit(lexemeComment)

	eatSpace(l)
	return lexStart
}

func lexSection(l *lexer_amod) stateFn {
	i := strings.Index(l.input[l.pos:], sectionModel)
	if i == 0 {
		l.pos += len(sectionModel)
		l.emit(lexemeSectionModel)
		eatSpace(l)
		return lexStart
	}

	i = strings.Index(l.input[l.pos:], sectionConfig)
	if i == 0 {
		l.pos += len(sectionConfig)
		l.emit(lexemeSectionConfig)
		eatSpace(l)
		return lexStart
	}

	i = strings.Index(l.input[l.pos:], sectionInit)
	if i == 0 {
		l.pos += len(sectionInit)
		l.emit(lexemeSectionInit)
		eatSpace(l)
		return lexStart
	}

	i = strings.Index(l.input[l.pos:], sectionProductions)
	if i == 0 {
		l.pos += len(sectionProductions)
		l.emit(lexemeSectionProductions)
		eatSpace(l)
		return lexStart
	}

	return lexSpace
}

func lexIdentifier(l *lexer_amod) stateFn {
	for {
		r := l.peek()

		if !isAlphaNumeric(r) {
			break
		}

		l.next()
	}

	// Perhaps not the best way to do this.
	// I'm sure there's a char-by-char way we could implement which would be faster.
	isKeyword := l.lookupKeyword(l.input[l.start:l.pos])
	if isKeyword {
		l.emit(lexemeKeyword)
	} else {
		l.emit(lexemeIdentifier)
	}

	return lexStart
}

func lexNumber(l *lexer_amod) stateFn {
	l.accept("+-")

	digits := "0123456789"

	l.acceptRun(digits)

	if l.accept(".") {
		l.acceptRun(digits)
	}

	l.emit(lexemeNumber)

	return lexStart
}

func lexQuotedString(l *lexer_amod) stateFn {
	quoteType := l.next()
	done := false

	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof:
			fallthrough
		case '\n':
			return l.errorf("unterminated quoted string")
		case quoteType:
			done = true
		}

		if done {
			break
		}
	}

	l.emit(lexemeString)

	return lexSpace
}

func lexDoItem(l *lexer_amod) stateFn {
	eatSpace(l)

	for {
		r := l.next()

		if isNewline(r) {
			break
		}
	}

	l.emit(lexemeDoCode)

	eatSpace(l)

	return lexDoBlock
}

func lexDoBlock(l *lexer_amod) stateFn {
	// check for ending ">#"
	if l.inDoBlock {
		r := l.next()
		if r == '>' {
			if l.peek() == '#' {
				l.next()
				l.inDoBlock = false
				l.emit(lexemeCodeEnd)

				return lexSpace
			}
		}
		l.backup()
	}

	return lexDoItem
}

// cleanData normalizes line endings
func cleanData(data *string) {
	*data = strings.Replace(*data, "\r\n", "\n", -1)
	*data = strings.Replace(*data, "\r", "\n", -1)
}
