package lexer

import (
	"github.com/defiant00/char/compiler/token"
	"github.com/defiant00/char/data"
	"unicode/utf8"
)

const (
	eof = -1
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

func (l *Lexer) emit(t token.Type) {
	l.tokens <- token.Token{t, l.lineCount(), l.current()}
	l.start = l.pos
	l.widths = data.NewStack(10)
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

// lexIndent lexes the initial indentation of a line
func lexIndent(l *Lexer) stateFn {
	indent := 0
	for {
		switch r := l.next(); r {
		case eof:
			l.discard()
			l.emitIndent(0)
			l.emit(token.EOF)
			return nil
		case '\r':
			l.discard()
		case '\n':
			l.emit(token.EOL)
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
