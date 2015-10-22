package token

import "fmt"

type Type int

func (t Type) String() string {
	return tStrings[t]
}

const (
	ERROR  Type = iota // an error, val contains the error text
	INDENT             // an increase in indentation
	DEDENT             // a decrease in indentation
	EOL                // the end of a line of code
	EOF                // the end of the file
	STRING             // a literal string
	CHAR               // a literal character
	NUMBER             // a literal number
	IDENT              // an identifier
)

var tStrings = map[Type]string{
	ERROR:  "Error",
	INDENT: "Indent",
	DEDENT: "Dedent",
	EOL:    "EOL",
	EOF:    "EOF",
	STRING: "String",
	CHAR:   "Character",
	NUMBER: "Number",
	IDENT:  "Identifier",
}

type Token struct {
	Type Type
	Pos  Position
	Val  string
}

func (t Token) String() string {
	switch t.Type {
	case EOL:
		return fmt.Sprintf("(%v) %v\n", t.Pos, t.Type)
	case STRING, CHAR, NUMBER, IDENT:
		return fmt.Sprintf("(%v) %v : '%v'", t.Pos, t.Type, t.Val)
	default:
		return fmt.Sprintf("(%v) %v", t.Pos, t.Type)
	}
}

type Position struct {
	Line, Char int
}

func (p Position) String() string {
	return fmt.Sprintf("%v:%v", p.Line, p.Char)
}
