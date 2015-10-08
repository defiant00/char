package cl

import (
	"fmt"
)

type parser struct {
	tokens []token // All tokens
	pos    int     // Current position in tokens
}

func (p *parser) next() token {
	n := p.peek()
	p.pos++
	return n
}

func (p *parser) backup() {
	p.pos--
}

func (p *parser) peek() token {
	return p.tokens[p.pos]
}

func (p *parser) parseProgram() exprAST {
	e := &programAST{}
	// Build the AST
loop:
	for {
		switch p.peek().typ {
		case tError, tEOF:
			break loop
		case tPackage:
			e.items = append(e.items, p.parsePackage())
		case tImport:
			e.items = append(e.items, p.parseImport())
		case tEOL:
			p.next()
		default:
			fmt.Println("\n\n*** Unknown token", p.peek())
			break loop
		}
	}
	return e
}

func (p *parser) parsePackage() exprAST {
	p.next() // Consume the package token
	t := p.next()
	if t.typ == tIdentifier {
		end := p.next()
		if end.isLineEnd() {
			return packageAST{name: t.val}
		}
		return errorAST{error: "Extra token found after package identifier."}
	}
	return errorAST{error: "No package identifier found!"}
}

func (p *parser) parseImport() exprAST {
	i := &importAST{}
	p.next() // Consume the import token
	read := true
	for read {
		t := p.next()
		switch t.typ {
		case tString: // Import string on same line as import command
			i.names = append(i.names, t.val)
			switch p.peek().typ {
			case tComma:
				p.next()
			case tEOL:
				p.next()
				read = false
			default:
				read = false
			}
		default:
			return errorAST{error: "Found invalid import."}
		}
	}
	return i
}

func parse(input string) exprAST {
	p := &parser{}
	l := lex(input)
	var t token

	// Read all non-comment tokens into a slice.
	for {
		t = l.nextToken()
		if t.typ != tSLComment {
			p.tokens = append(p.tokens, t)
		}
		fmt.Print(" ", t)
		if t.typ == tError || t.typ == tEOF {
			break
		}
	}

	return p.parseProgram()
}
