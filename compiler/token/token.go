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

func (t Type) IsInBlock() bool {
	return t == COMMA || t == RIGHTPAREN
}

const (
	ERROR         Type = iota // an error, val contains the error text
	INDENT                    // an increase in indentation
	DEDENT                    // a decrease in indentation
	EOL                       // the end of a line of code
	EOF                       // the end of the file
	COMMENT                   // comment
	STRING                    // a literal string
	CHAR                      // a literal character
	NUMBER                    // a literal number
	IDENTIFIER                // an identifier
	keyword_start             //
	USE                       // 'use'
	AS                        // 'as'
	IS                        // 'is'
	WITH                      // 'with'
	FUNCTION                  // 'fn'
	MIXIN                     // 'mix'
	VAR                       // 'var'
	BLANK                     // '_'
	RETURN                    // 'ret'
	DEFER                     // 'defer'
	IOTA                      // 'iota'
	TRUE                      // 'true'
	FALSE                     // 'false'
	EQUALS                    // '=='
	NOTEQUALS                 // '!='
	AND                       // 'and'
	OR                        // 'or'
	DOT                       // '.'
	COMMA                     // ','
	LEFTPAREN                 // '('
	RIGHTPAREN                // ')'
	LEFTCARET                 // '<'
	RIGHTCARET                // '>'
	ASSIGN                    // '='
	ADDASSIGN                 // '+='
	SUBASSIGN                 // '-='
	MULASSIGN                 // '*='
	DIVASSIGN                 // '/='
	MODASSIGN                 // '%='
	ADD                       // '+'
	SUB                       // '-'
	MUL                       // '*'
	DIV                       // '/'
	MOD                       // '%'
	keyword_end               //
)

var tStrings = map[Type]string{
	ERROR:      "error",
	INDENT:     "indent",
	DEDENT:     "dedent",
	EOL:        "EOL",
	EOF:        "EOF",
	COMMENT:    "comment",
	STRING:     "string",
	CHAR:       "char",
	NUMBER:     "number",
	IDENTIFIER: "id",
	USE:        "use",
	AS:         "as",
	IS:         "is",
	WITH:       "with",
	FUNCTION:   "fn",
	MIXIN:      "mix",
	VAR:        "var",
	BLANK:      "_",
	RETURN:     "ret",
	DEFER:      "defer",
	IOTA:       "iota",
	TRUE:       "true",
	FALSE:      "false",
	EQUALS:     "==",
	NOTEQUALS:  "!=",
	AND:        "and",
	OR:         "or",
	DOT:        ".",
	COMMA:      ",",
	LEFTPAREN:  "(",
	RIGHTPAREN: ")",
	LEFTCARET:  "<",
	RIGHTCARET: ">",
	ASSIGN:     "=",
	ADDASSIGN:  "+=",
	SUBASSIGN:  "-=",
	MULASSIGN:  "*=",
	DIVASSIGN:  "/=",
	MODASSIGN:  "%=",
	ADD:        "+",
	SUB:        "-",
	MUL:        "*",
	DIV:        "/",
	MOD:        "%",
}

var Keywords = map[string]Type{
	"use":   USE,
	"as":    AS,
	"is":    IS,
	"with":  WITH,
	"fn":    FUNCTION,
	"mix":   MIXIN,
	"var":   VAR,
	"_":     BLANK,
	"ret":   RETURN,
	"defer": DEFER,
	"iota":  IOTA,
	"true":  TRUE,
	"false": FALSE,
	"==":    EQUALS,
	"!=":    NOTEQUALS,
	"and":   AND,
	"or":    OR,
	".":     DOT,
	",":     COMMA,
	"(":     LEFTPAREN,
	")":     RIGHTPAREN,
	"<":     LEFTCARET,
	">":     RIGHTCARET,
	"=":     ASSIGN,
	"+=":    ADDASSIGN,
	"-=":    SUBASSIGN,
	"*=":    MULASSIGN,
	"/=":    DIVASSIGN,
	"%=":    MODASSIGN,
	"+":     ADD,
	"-":     SUB,
	"*":     MUL,
	"/":     DIV,
	"%":     MOD,
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
	case COMMENT, STRING, CHAR, NUMBER, IDENTIFIER, ERROR:
		return fmt.Sprintf("%v %v : '%v'", t.Pos, t.Type, t.Val)
	default:
		return fmt.Sprintf("%v %v", t.Pos, t.Type)
	}
}

func (t Token) Precedence() int {
	switch t.Type {
	case ASSIGN, ADDASSIGN, SUBASSIGN, MULASSIGN, DIVASSIGN, MODASSIGN:
		return 1
	case COMMA:
		return 2
	case ADD, SUB:
		return 3
	case MUL, DIV, MOD:
		return 4
	case EQUALS, NOTEQUALS:
		return 5
	case AND, OR, IS:
		return 6
	case LEFTCARET, RIGHTCARET:
		return 7
	case AS:
		return 8
	case DOT:
		return 10
	}
	return -1
}

type Position struct {
	Line, Char int
}

func (p Position) String() string {
	return fmt.Sprintf("%v:%v", p.Line, p.Char)
}
