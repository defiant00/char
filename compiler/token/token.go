package token

import "fmt"

type Type int

func (t Type) String() string {
	return tStrings[t]
}

func (t Type) IsKeyword() bool {
	return t > keyword_start && t < keyword_end
}

func (t Type) IsType() bool {
	return t == IDENTIFIER || t == FUNCTION
}

const (
	ERROR         Type = iota // an error, val contains the error text
	INDENT                    // an increase in indentation
	DEDENT                    // a decrease in indentation
	EOL                       // the end of a line of code
	EOF                       // the end of the file
	SLCOMMENT                 // single-line comment
	STRING                    // a literal string
	CHAR                      // a literal character
	NUMBER                    // a literal number
	IDENTIFIER                // an identifier
	keyword_start             //
	USE                       // 'use'
	AS                        // 'as'
	WITH                      // 'with'
	FUNCTION                  // 'func'
	VAR                       // 'var'
	IOTA                      // 'iota'
	DOT                       // '.'
	COMMA                     // ','
	LEFTPAREN                 // '('
	RIGHTPAREN                // ')'
	LEFTCARET                 // '<'
	RIGHTCARET                // '>'
	ASSIGN                    // '='
	ADD                       // '+'
	SUBTRACT                  // '-'
	MULTIPLY                  // '*'
	DIVIDE                    // '/'
	MOD                       // '%'
	keyword_end               //
)

var tStrings = map[Type]string{
	ERROR:      "Error",
	INDENT:     "Indent",
	DEDENT:     "Dedent",
	EOL:        "EOL",
	EOF:        "EOF",
	SLCOMMENT:  "SLComment",
	STRING:     "String",
	CHAR:       "Character",
	NUMBER:     "Number",
	IDENTIFIER: "Identifier",
	USE:        "Use",
	AS:         "As",
	WITH:       "With",
	FUNCTION:   "Func",
	VAR:        "Var",
	IOTA:       "Iota",
	DOT:        ".",
	COMMA:      ",",
	LEFTPAREN:  "(",
	RIGHTPAREN: ")",
	LEFTCARET:  "<",
	RIGHTCARET: ">",
	ASSIGN:     "=",
	ADD:        "+",
	SUBTRACT:   "-",
	MULTIPLY:   "*",
	DIVIDE:     "/",
	MOD:        "%",
}

var Keywords = map[string]Type{
	"use":  USE,
	"as":   AS,
	"with": WITH,
	"func": FUNCTION,
	"var":  VAR,
	"iota": IOTA,
	".":    DOT,
	",":    COMMA,
	"(":    LEFTPAREN,
	")":    RIGHTPAREN,
	"<":    LEFTCARET,
	">":    RIGHTCARET,
	"=":    ASSIGN,
	"+":    ADD,
	"-":    SUBTRACT,
	"*":    MULTIPLY,
	"/":    DIVIDE,
	"%":    MOD,
}

type Token struct {
	Type Type
	Pos  Position
	Val  string
}

func (t Token) String() string {
	switch t.Type {
	case EOL:
		return fmt.Sprintf("%v %v\n", t.Pos, t.Type)
	case SLCOMMENT, STRING, CHAR, NUMBER, IDENTIFIER, ERROR:
		return fmt.Sprintf("%v %v : '%v'", t.Pos, t.Type, t.Val)
	default:
		return fmt.Sprintf("%v %v", t.Pos, t.Type)
	}
}

type Position struct {
	Line, Char int
}

func (p Position) String() string {
	return fmt.Sprintf("%v:%v", p.Line, p.Char)
}
