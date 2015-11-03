package lexer

import (
	"fmt"
	"github.com/defiant00/char/compiler/token"
	"github.com/defiant00/char/data"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	eof           = -1
	operatorChars = "()[]<>{}!=+-*/%,._:&|^"
)

type stateFn func(*Lexer) stateFn

type Lexer struct {
	input        string           // The string being scanned
	state        stateFn          // The current lexer state function
	indentLevels *data.Stack      // A stack of the current indentation levels
	start        int              // Start position of this token
	pos          int              // Current position in the input
	widths       *data.Stack      // Width of the runes read from the stack since the last emit
	tokens       chan token.Token // Channel of scanned tokens
	inStmt       bool             // Whether we are currently in a statement
}

func (l *Lexer) next() rune {
	if l.pos >= len(l.input) {
		l.widths.Push(0)
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.widths.Push(w)
	l.pos += w
	return r
}

func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *Lexer) current() string {
	return l.input[l.start:l.pos]
}

func (l *Lexer) backup() {
	l.pos -= l.widths.Pop()
}

func (l *Lexer) discard() {
	l.start = l.pos
}

func (l *Lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *Lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- token.Token{token.ERROR, token.Position{l.lineCount(), l.charCount()}, fmt.Sprintf(format, args...)}
	return nil
}

func (l *Lexer) emit(t token.Type) {
	l.tokens <- token.Token{t, token.Position{l.lineCount(), l.charCount()}, l.current()}
	l.start = l.pos
	l.widths = data.NewStack(10)
}

func (l *Lexer) emitIndent(indent int) {
	i := l.indentLevels.Peek()
	if indent > i {
		l.emit(token.INDENT)
		l.indentLevels.Push(indent)
	} else {
		for l.indentLevels.Len() > 0 && indent < i {
			l.emit(token.DEDENT)
			l.emit(token.EOL)
			l.indentLevels.Pop()
			i = l.indentLevels.Peek()
		}
		if l.indentLevels.Len() == 0 || i != indent {
			l.errorf("Mismatched indentation level encountered.")
		}
	}
}

func (l *Lexer) lineCount() int {
	return strings.Count(l.input[:l.start], "\n") + 1
}

func (l *Lexer) charCount() int {
	c := 1
	for i := l.start - 1; i > -1 && l.input[i] != '\n'; i-- {
		c++
	}
	return c
}

func (l *Lexer) run() {
	for l.state != nil {
		l.state = l.state(l)
	}
	close(l.tokens)
}

func Lex(input string) *Lexer {
	l := &Lexer{
		input:        input,
		state:        lexIndent,
		indentLevels: data.NewStack(10),
		widths:       data.NewStack(10),
		tokens:       make(chan token.Token, 10),
	}
	l.indentLevels.Push(0)
	go l.run()
	return l
}

func (l *Lexer) NextToken() token.Token {
	return <-l.tokens
}

// lexIndent lexes the initial indentation of a line
func lexIndent(l *Lexer) stateFn {
	l.inStmt = false
	indent := 0
	for {
		switch r := l.next(); r {
		case eof:
			l.discard()
			l.emitIndent(0)
			l.emit(token.EOF)
			return nil
		case '\r', '\n':
			indent = 0
			l.discard()
		case ' ':
			indent++
		case '\t':
			indent += 4
		case ';':
			l.backup()
			l.discard()
			return lexComment
		default:
			l.backup()
			l.discard()
			l.emitIndent(indent)
			return lexStatement
		}
	}
}

// lexStatement lexes general statements into identifiers, symbols and literals
func lexStatement(l *Lexer) stateFn {
	for {
		switch r := l.peek(); {
		case r == eof:
			if l.inStmt {
				l.emit(token.EOL)
			}
			l.emitIndent(0)
			l.emit(token.EOF)
		case r == ' ' || r == '\t' || r == '\r':
			l.next()
			l.discard()
		case r == '\n':
			l.next()
			if l.inStmt {
				l.emit(token.EOL)
			}
			return lexIndent
		case r == ';':
			return lexComment
		case r == '"':
			l.inStmt = true
			return lexString
		case r == '\'':
			l.inStmt = true
			return lexChar
		case unicode.IsLetter(r):
			l.inStmt = true
			return lexIdentifier
		case unicode.IsDigit(r):
			l.inStmt = true
			return lexNumber
		case l.accept(operatorChars):
			l.inStmt = true
			return lexOperator
		default:
			l.errorf("Invalid rune %c encountered.", r)
		}
	}
}

func lexComment(l *Lexer) stateFn {
	l.next() // Eat the ;
	l.discard()
	for r := l.peek(); r != eof && r != '\r' && r != '\n'; {
		l.next()
		r = l.peek()
	}
	l.emit(token.COMMENT)
	return lexStatement
}

func lexIdentifier(l *Lexer) stateFn {
	for r := l.peek(); unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'; {
		l.next()
		r = l.peek()
	}
	t := token.Keywords[l.current()]
	if !t.IsKeyword() {
		t = token.IDENTIFIER
	}
	l.emit(t)
	return lexStatement
}

func lexNumber(l *Lexer) stateFn {
	for r := l.peek(); unicode.IsDigit(r); {
		l.next()
		r = l.peek()
	}
	l.accept(".")
	for r := l.peek(); unicode.IsDigit(r); {
		l.next()
		r = l.peek()
	}
	l.emit(token.NUMBER)
	return lexStatement
}

func lexOperator(l *Lexer) stateFn {
	l.acceptRun(operatorChars)
	p := l.pos
	t := token.Keywords[l.current()]
	for !t.IsKeyword() && l.pos > l.start {
		l.backup()
		t = token.Keywords[l.current()]
	}
	if l.pos > l.start {
		l.emit(t)
		return lexStatement
	}
	return l.errorf("Invalid operator '%v'", l.input[l.start:p])
}

func lexString(l *Lexer) stateFn {
	l.next()
	l.discard()
	inEsc := false
	for {
		switch r := l.next(); {
		case r == eof || r == '\r' || r == '\n':
			return l.errorf("Unclosed \"")
		case !inEsc && r == '\\':
			inEsc = true
		case !inEsc && r == '"':
			l.backup()
			l.emit(token.STRING)
			l.next()
			l.discard()
			return lexStatement
		case inEsc:
			inEsc = false
		}
	}
}

// lexChar lexes a literal character
func lexChar(l *Lexer) stateFn {
	l.next()
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
				l.emit(token.CHAR)
				l.next()
				l.discard()
				return lexStatement
			}
			return l.errorf("Unclosed '")
		}
	default:
		if l.accept("'") {
			l.backup()
			l.emit(token.CHAR)
			l.next()
			l.discard()
			return lexStatement
		}
		return l.errorf("Unclosed '")
	}
}
