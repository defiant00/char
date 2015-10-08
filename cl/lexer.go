package cl

import (
	"fmt"
	"github.com/defiant00/char/data"
	"strings"
	"unicode/utf8"
)

const (
	eof           = -1
	lineComment   = "//"
	letters       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers       = "0123456789"
	alphaNumeric  = letters + numbers
	operatorChars = "()!=<>+-*/%,._"
)

type tType int

const (
	tError      tType = iota // An error, val contains the error text
	tIndent                  // An increates in indentation
	tDedent                  // A decrease in indentation
	tEOL                     // End of line
	tEOF                     // End of file
	tSLComment               // Single-line comment starting with '//'
	tString                  // Literal string
	tChar                    // Literal character
	tNumber                  // Literal number
	tIdentifier              // Identifier
	tKeyword                 // Everything below this token is a keyword
	tPackage                 // 'package'
	tImport                  // 'import'
	tVar                     // 'var'
	tAs                      // 'as'
	tNil                     // 'nil'
	tTrue                    // 'true'
	tFalse                   // 'false'
	tAnd                     // 'and'
	tOr                      // 'or'
	tDot                     // '.'
	tComma                   // ','
	tLeftParen               // '('
	tRightParen              // ')'
	tAssign                  // '='
	tAdd                     // '+'
	tSubtract                // '-'
	tMultiply                // '*'
	tDivide                  // '/'
	tMod                     // '%'
	tAny                     // '_'
)

var tStrings = map[tType]string{
	tError:      "Error",
	tIndent:     "Indent",
	tDedent:     "Dedent",
	tEOL:        "EOL",
	tEOF:        "EOF",
	tSLComment:  "SLComment",
	tString:     "String",
	tChar:       "Char",
	tNumber:     "Number",
	tIdentifier: "Identifier",
	tKeyword:    "Keyword",
	tPackage:    "Package",
	tImport:     "Import",
	tVar:        "Var",
	tAs:         "As",
	tNil:        "Nil",
	tTrue:       "True",
	tFalse:      "False",
	tAnd:        "And",
	tOr:         "Or",
	tDot:        "Dot",
	tComma:      "Comma",
	tLeftParen:  "LeftParen",
	tRightParen: "RightParen",
	tAssign:     "Assign",
	tAdd:        "Add",
	tSubtract:   "Subtract",
	tMultiply:   "Multiply",
	tDivide:     "Divide",
	tMod:        "Mod",
	tAny:        "Any",
}

func (t tType) String() string {
	return tStrings[t]
}

var opKeywords = map[string]tType{
	".": tDot,
	",": tComma,
	"(": tLeftParen,
	")": tRightParen,
	"=": tAssign,
	"+": tAdd,
	"-": tSubtract,
	"*": tMultiply,
	"/": tDivide,
	"%": tMod,
	"_": tAny,
}

var resKeywords = map[string]tType{
	"package": tPackage,
	"import":  tImport,
	"var":     tVar,
	"as":      tAs,
	"nil":     tNil,
	"true":    tTrue,
	"false":   tFalse,
	"and":     tAnd,
	"or":      tOr,
}

type token struct {
	typ  tType
	line int
	val  string
}

func (t token) isLineEnd() bool {
	return t.typ == tEOL || t.typ == tEOF
}

func (t token) String() string {
	switch t.typ {
	case tEOL:
		return fmt.Sprintf("(%v) %v\n", t.line, t.typ)
	case tEOF, tIndent, tDedent, tDot, tComma, tLeftParen, tRightParen, tAssign, tAdd, tSubtract, tMultiply, tDivide, tMod, tAny, tPackage, tImport, tVar, tAs, tNil, tTrue, tFalse, tAnd, tOr:
		return fmt.Sprintf("(%v) %v", t.line, t.typ)
	default:
		return fmt.Sprintf("(%v) %v : '%v'", t.line, t.typ, t.val)
	}
}

type lexer struct {
	input        string      // The string being scanned
	state        stateFn     // The current lexer state function
	indentLevels *data.Stack // A stack of the current indentation levels
	start        int         // Start position of this token
	pos          int         // Current position in the input
	widths       *data.Stack // Width of the runes read from the stack since the last emit
	tokens       chan token  // Channel of scanned tokens
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.widths.Push(0)
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.widths.Push(w)
	l.pos += w
	return r
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) current() string {
	return l.input[l.start:l.pos]
}

func (l *lexer) backup() {
	l.pos -= l.widths.Pop()
}

func (l *lexer) discard() {
	l.start = l.pos
}

func (l *lexer) emit(t tType) {
	l.tokens <- token{t, l.lineCount(), l.current()}
	l.start = l.pos
	l.widths = data.NewStack(10)
}

func (l *lexer) emitIndent(indent int) {
	i := l.indentLevels.Peek()
	if indent > i {
		l.emit(tIndent)
		l.indentLevels.Push(indent)
	} else {
		for l.indentLevels.Len() > 0 && indent < i {
			l.emit(tDedent)
			l.indentLevels.Pop()
			i = l.indentLevels.Peek()
		}
		if l.indentLevels.Len() == 0 || i != indent {
			l.errorf("Mismatched indentation level encountered.")
		}
	}
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- token{tError, l.lineCount(), fmt.Sprintf(format, args...)}
	return nil
}

func (l *lexer) nextToken() token {
	return <-l.tokens
}

func (l *lexer) startsWith(start string) bool {
	return strings.HasPrefix(l.input[l.pos:], start)
}

func (l *lexer) lineCount() int {
	return 1 + strings.Count(l.input[:l.start], "\n")
}

func (l *lexer) run() {
	for l.state != nil {
		l.state = l.state(l)
	}
	close(l.tokens)
}

func lex(input string) *lexer {
	l := &lexer{
		input:        input,
		state:        lexIndent,
		indentLevels: data.NewStack(10),
		widths:       data.NewStack(10),
		tokens:       make(chan token, 10),
	}
	l.indentLevels.Push(0)
	go l.run()
	return l
}

type stateFn func(*lexer) stateFn

// lexIndent lexes the initial indentation of a line.
func lexIndent(l *lexer) stateFn {
	indent := 0
	for {
		if l.startsWith(lineComment) {
			l.discard()
			return lexSLComment
		}
		switch r := l.next(); r {
		case eof:
			l.discard()
			l.emitIndent(0)
			l.emit(tEOF)
			return nil
		case '\r':
			l.discard()
		case '\n':
			l.emit(tEOL)
			return lexIndent
		case ' ':
			indent++
		case '\t':
			indent += 4
		default:
			l.backup()
			l.discard()
			l.emitIndent(indent)
			return lexStatement
		}
	}
}

// lexStatement lexes general statements into identifiers, symbols and literals
func lexStatement(l *lexer) stateFn {
	for {
		switch {
		case l.startsWith(lineComment):
			return lexSLComment
		case l.accept(letters):
			return lexIdentifier
		case l.accept(numbers):
			l.backup()
			return lexNumber
		case l.accept(operatorChars):
			return lexOperator
		default:
			switch r := l.next(); r {
			case eof:
				l.emitIndent(0)
				l.emit(tEOF)
				return nil
			case '\n':
				l.emit(tEOL)
				return lexIndent
			case ' ', '\t', '\r':
				l.discard()
			case '\'':
				return lexChar
			case '"':
				return lexString
			default:
				return l.errorf("Invalid rune '%v' encountered.", l.current())
			}
		}
	}
}

// lexOperator lexes an operator, the first character has already been consumed
func lexOperator(l *lexer) stateFn {
	l.acceptRun(operatorChars)
	p := l.pos
	t := opKeywords[l.current()]
	for t < tKeyword && l.pos > l.start {
		l.backup()
		t = opKeywords[l.current()]
	}
	if l.pos > l.start {
		l.emit(t)
		return lexStatement
	}
	return l.errorf("Invalid operator '%v'", l.input[l.start:p])
}

// lexNumber lexes a number
func lexNumber(l *lexer) stateFn {
	l.acceptRun(numbers)
	l.acceptRun(".")
	l.acceptRun(numbers)
	l.emit(tNumber)
	return lexStatement
}

// lexIdentifier lexes an identifier, anything that starts with a letter
// and contains only letters and numbers. The first character has already been
// consumed
func lexIdentifier(l *lexer) stateFn {
	l.acceptRun(alphaNumeric)
	t := resKeywords[l.current()]
	if t < tKeyword {
		t = tIdentifier
	}
	l.emit(t)
	return lexStatement
}

// lexSLComment lexes a single line comment, starting with //
func lexSLComment(l *lexer) stateFn {
	for {
		switch r := l.next(); r {
		case eof:
			l.emit(tSLComment)
			l.emitIndent(0)
			l.emit(tEOF)
			return nil
		case '\r', '\n':
			l.backup()
			l.emit(tSLComment)
			return lexStatement
		}
	}
}

// lexString lexes a literal string, with the opening " already consumed
func lexString(l *lexer) stateFn {
	l.discard()
	inEsc := false
	for {
		switch r := l.next(); {
		case r == eof || r == '\n':
			return l.errorf("Unclosed \"")
		case !inEsc && r == '\\':
			inEsc = true
		case !inEsc && r == '"':
			l.backup()
			l.emit(tString)
			l.next()
			l.discard()
			return lexStatement
		case inEsc:
			inEsc = false
		}
	}
}

// lexChar lexes a literal character, with the opening ' aleady consumed
func lexChar(l *lexer) stateFn {
	l.discard()
	switch r := l.next(); r {
	case eof, '\r', '\n':
		return l.errorf("Unclosed '")
	case '\\':
		switch r2 := l.next(); r2 {
		case eof, '\r', '\n':
			return l.errorf("Unclosed '")
		default:
			if l.accept("'") {
				l.backup()
				l.emit(tChar)
				l.next()
				l.discard()
				return lexStatement
			}
			return l.errorf("Unclosed '")
		}
	default:
		if l.accept("'") {
			l.backup()
			l.emit(tChar)
			l.next()
			l.discard()
			return lexStatement
		}
		return l.errorf("Unclosed '")
	}
}