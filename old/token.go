package cl

import (
	"fmt"
)

type tType int

const (
	tError      tType = iota // An error, val contains the error text
	tIndent                  // An increates in indentation
	tDedent                  // A decrease in indentation
	tEOL                     // End of line
	tEOF                     // End of file
	tGoBlock                 // A block of Go code
	tSLComment               // Single-line comment starting with '//'
	tString                  // Literal string
	tChar                    // Literal character
	tNumber                  // Literal number
	tIdentifier              // Identifier
	tKeyword                 // Everything below this token is a keyword
	tUse                     // 'use'
	tVar                     // 'var'
	tReturn                  // 'return'
	tLeftParen               // '('
	tRightParen              // ')'
	tDot                     // '.'
	tComma                   // ','
	tAssign                  // '='
	tAddAssign               // '+='
	tAdd                     // '+'
	tMultiply                // '*'
	tEquals                  // '=='
	tAnd                     // 'and'
	tOr                      // 'or'
	tTrue                    // 'true'
	tFalse                   // 'false'
)

var tStrings = map[tType]string{
	tError:      "Error",
	tIndent:     "Indent",
	tDedent:     "Dedent",
	tEOL:        "EOL",
	tEOF:        "EOF",
	tGoBlock:    "GoBlock",
	tSLComment:  "SLComment",
	tString:     "String",
	tChar:       "Char",
	tNumber:     "Number",
	tIdentifier: "Identifier",
	tKeyword:    "Keyword",
	tUse:        "Use",
	tVar:        "Var",
	tReturn:     "Return",
	tLeftParen:  "LeftParen",
	tRightParen: "RightParen",
	tDot:        "Dot",
	tComma:      "Comma",
	tAssign:     "Assign",
	tAddAssign:  "AddAssign",
	tAdd:        "Add",
	tMultiply:   "Multiply",
	tEquals:     "Equals",
	tAnd:        "And",
	tOr:         "Or",
	tTrue:       "True",
	tFalse:      "False",
}

var tPrecedences = map[tType]int{
	tAssign:    10,
	tAddAssign: 10,
	tComma:     20,
	tAdd:       30,
	tMultiply:  40,
	tAnd:       70,
	tOr:        80,
	tEquals:    90,
	tDot:       100,
}

func (t tType) String() string {
	return tStrings[t]
}

var tokGoOp = map[tType]string{
	tDot:       ".",
	tComma:     ",",
	tAssign:    "=",
	tAddAssign: "+=",
	tAdd:       "+",
	tMultiply:  "*",
	tEquals:    "==",
	tAnd:       "&&",
	tOr:        "||",
}

var opKeywords = map[string]tType{
	"(":  tLeftParen,
	")":  tRightParen,
	".":  tDot,
	",":  tComma,
	"=":  tAssign,
	"+=": tAddAssign,
	"+":  tAdd,
	"*":  tMultiply,
	"==": tEquals,
}

var resKeywords = map[string]tType{
	"use":    tUse,
	"var":    tVar,
	"return": tReturn,
	"and":    tAnd,
	"or":     tOr,
	"true":   tTrue,
	"false":  tFalse,
}

type token struct {
	typ  tType
	line int
	val  string
}

func (t token) String() string {
	switch t.typ {
	case tEOL:
		return fmt.Sprintf("(%v) %v\n", t.line, t.typ)
	case tEOF, tIndent, tDedent, tLeftParen, tRightParen, tDot, tComma, tAssign, tAddAssign, tAdd, tMultiply, tUse, tVar, tReturn, tEquals, tAnd, tOr, tTrue, tFalse:
		return fmt.Sprintf("(%v) %v", t.line, t.typ)
	default:
		return fmt.Sprintf("(%v) %v : '%v'", t.line, t.typ, t.val)
	}
}

func (t token) precedence() int {
	p := tPrecedences[t.typ]
	if p <= 0 {
		return -1
	}
	return p
}
