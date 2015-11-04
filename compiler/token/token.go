package token

import "fmt"

type Type int

func (t Type) String() string {
	return tStrings[t]
}

func (t Type) IsKeyword() bool {
	return t > keyword_start && t < keyword_end
}

func (t Type) IsUnaryOp() bool {
	return t > unary_op_start && t < unary_op_end
}

func (t Type) IsAssign() bool {
	return t > assign_start && t < assign_end
}

func (t Type) IsType() bool {
	return t == IDENTIFIER || t == FUNCTION
}

func (t Type) IsInBlock() bool {
	return t == COMMA || t == RIGHT_PAREN
}

const (
	ERROR          Type = iota // an error, val contains the error text
	INDENT                     // an increase in indentation
	DEDENT                     // a decrease in indentation
	EOL                        // the end of a line of code
	EOF                        // the end of the file
	COMMENT                    // comment
	STRING                     // a literal string
	CHAR                       // a literal character
	NUMBER                     // a literal number
	IDENTIFIER                 // an identifier
	keyword_start              //
	USE                        // 'use'
	AS                         // 'as'
	IF                         // 'if'
	IS                         // 'is'
	IN                         // 'in'
	WITH                       // 'with'
	FUNCTION                   // 'fn'
	INTERFACE                  // 'intf'
	MIXIN                      // 'mix'
	VAR                        // 'var'
	BLANK                      // '_'
	RETURN                     // 'ret'
	DEFER                      // 'defer'
	FOR                        // 'for'
	LOOP                       // 'loop'
	BREAK                      // 'break'
	IOTA                       // 'iota'
	TRUE                       // 'true'
	FALSE                      // 'false'
	EQUAL                      // '=='
	NOT_EQUAL                  // '!='
	LEFT_CARET                 // '<'
	RIGHT_CARET                // '>'
	LT_EQUAL                   // '<='
	GT_EQUAL                   // '>='
	AND                        // 'and'
	OR                         // 'or'
	DOT                        // '.'
	COMMA                      // ','
	COLON                      // ':'
	LEFT_PAREN                 // '('
	RIGHT_PAREN                // ')'
	LEFT_BRACKET               // '['
	RIGHT_BRACKET              // ']'
	LEFT_CURLY                 // '{'
	RIGHT_CURLY                // '}'
	assign_start               //
	ASSIGN                     // '='
	ADD_ASSIGN                 // '+='
	SUB_ASSIGN                 // '-='
	MUL_ASSIGN                 // '*='
	DIV_ASSIGN                 // '/='
	MOD_ASSIGN                 // '%='
	B_AND_ASSIGN               // '&='
	B_OR_ASSIGN                // '|='
	B_XOR_ASSIGN               // '^='
	LSHIFT_ASSIGN              // '<<='
	RSHIFT_ASSIGN              // '>>='
	assign_end                 //
	ADD                        // '+'
	MUL                        // '*'
	DIV                        // '/'
	MOD                        // '%'
	LSHIFT                     // '<<'
	RSHIFT                     // '>>'
	B_AND                      // '&'
	B_OR                       // '|'
	B_XOR                      // '^'
	unary_op_start             //
	SUB                        // '-'
	NOT                        // '!'
	unary_op_end               //
	keyword_end                //
)

var tStrings = map[Type]string{
	ERROR:         "error",
	INDENT:        "indent",
	DEDENT:        "dedent",
	EOL:           "EOL",
	EOF:           "EOF",
	COMMENT:       "comment",
	STRING:        "string",
	CHAR:          "char",
	NUMBER:        "number",
	IDENTIFIER:    "id",
	USE:           "use",
	AS:            "as",
	IF:            "if",
	IS:            "is",
	IN:            "in",
	WITH:          "with",
	FUNCTION:      "fn",
	INTERFACE:     "intf",
	MIXIN:         "mix",
	VAR:           "var",
	BLANK:         "_",
	RETURN:        "ret",
	DEFER:         "defer",
	FOR:           "for",
	LOOP:          "loop",
	BREAK:         "break",
	IOTA:          "iota",
	TRUE:          "true",
	FALSE:         "false",
	EQUAL:         "==",
	NOT_EQUAL:     "!=",
	LEFT_CARET:    "<",
	RIGHT_CARET:   ">",
	LT_EQUAL:      "<=",
	GT_EQUAL:      ">=",
	AND:           "and",
	OR:            "or",
	DOT:           ".",
	COMMA:         ",",
	COLON:         ":",
	LEFT_PAREN:    "(",
	RIGHT_PAREN:   ")",
	LEFT_BRACKET:  "[",
	RIGHT_BRACKET: "]",
	LEFT_CURLY:    "{",
	RIGHT_CURLY:   "}",
	ASSIGN:        "=",
	ADD_ASSIGN:    "+=",
	SUB_ASSIGN:    "-=",
	MUL_ASSIGN:    "*=",
	DIV_ASSIGN:    "/=",
	MOD_ASSIGN:    "%=",
	B_AND_ASSIGN:  "&=",
	B_OR_ASSIGN:   "|=",
	B_XOR_ASSIGN:  "^=",
	LSHIFT_ASSIGN: "<<=",
	RSHIFT_ASSIGN: ">>=",
	ADD:           "+",
	SUB:           "-",
	MUL:           "*",
	DIV:           "/",
	MOD:           "%",
	LSHIFT:        "<<",
	RSHIFT:        ">>",
	B_AND:         "&",
	B_OR:          "|",
	B_XOR:         "^",
	NOT:           "!",
}

var Keywords = map[string]Type{
	"use":   USE,
	"as":    AS,
	"if":    IF,
	"is":    IS,
	"in":    IN,
	"with":  WITH,
	"fn":    FUNCTION,
	"intf":  INTERFACE,
	"mix":   MIXIN,
	"var":   VAR,
	"_":     BLANK,
	"ret":   RETURN,
	"defer": DEFER,
	"for":   FOR,
	"loop":  LOOP,
	"break": BREAK,
	"iota":  IOTA,
	"true":  TRUE,
	"false": FALSE,
	"==":    EQUAL,
	"!=":    NOT_EQUAL,
	"<":     LEFT_CARET,
	">":     RIGHT_CARET,
	"<=":    LT_EQUAL,
	">=":    GT_EQUAL,
	"and":   AND,
	"or":    OR,
	".":     DOT,
	",":     COMMA,
	":":     COLON,
	"(":     LEFT_PAREN,
	")":     RIGHT_PAREN,
	"[":     LEFT_BRACKET,
	"]":     RIGHT_BRACKET,
	"{":     LEFT_CURLY,
	"}":     RIGHT_CURLY,
	"=":     ASSIGN,
	"+=":    ADD_ASSIGN,
	"-=":    SUB_ASSIGN,
	"*=":    MUL_ASSIGN,
	"/=":    DIV_ASSIGN,
	"%=":    MOD_ASSIGN,
	"&=":    B_AND_ASSIGN,
	"|=":    B_OR_ASSIGN,
	"^=":    B_XOR_ASSIGN,
	"<<=":   LSHIFT_ASSIGN,
	">>=":   RSHIFT_ASSIGN,
	"+":     ADD,
	"-":     SUB,
	"*":     MUL,
	"/":     DIV,
	"%":     MOD,
	"<<":    LSHIFT, // RSHIFT purposefully excluded since it clashes
	"&":     B_AND,  // with generics. See parser.peekCombo and
	"|":     B_OR,   // parser.nextCombo for where this is handled.
	"^":     B_XOR,
	"!":     NOT,
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
	case DOT:
		return 7
	case AS, IS:
		return 6
	case MUL, DIV, MOD, LSHIFT, RSHIFT, B_AND:
		return 5
	case ADD, SUB, B_OR, B_XOR:
		return 4
	case EQUAL, NOT_EQUAL, LEFT_CARET, LT_EQUAL, RIGHT_CARET, GT_EQUAL:
		return 3
	case AND:
		return 2
	case OR:
		return 1
	}
	return -1
}

type Position struct {
	Line, Char int
}

func (p Position) String() string {
	return fmt.Sprintf("%v:%v", p.Line, p.Char)
}
