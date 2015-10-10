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
			e.add(p.parsePackage())
		case tGoBlock:
			e.add(p.parseGoBlock())
		case tVar:
			e.add(p.parseVariable())
		case tNumber:
			e.add(p.parseNumber())
		case tEOL:
			p.next()
		default:
			fmt.Println("\n\n*** Unknown token", p.peek())
			break loop
		}
	}
	return e
}

func (p *parser) parseExpr() exprAST {
	switch p.peek().typ {
	case tVar:
		return p.parseVariable()
	case tNumber:
		return p.parseNumber()
	default:
		return errorAST{error: "Found something other than an expression."}
	}
}

func (p *parser) parseParenExpr() exprAST {
	p.next() // Consume '('
	e := p.parseExpr()

	// todo - check for an error?

	if p.peek().typ != tRightParen {
		return errorAST{error: "Missing )"}
	}
	p.next() // Consume ')'
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

func (p *parser) parseVariable() exprAST {
	p.next() // Consume the 'var'
	t := p.next()
	if t.typ == tIdentifier {
		end := p.next()
		if end.isLineEnd() {
			return variableAST{name: t.val}
		}
		return errorAST{error: "Var found with more stuff on the line."}
	}
	return errorAST{error: "Var found with no variable name specified."}
}

func (p *parser) parseNumber() exprAST {
	return numberAST{number: p.next().val}
}

func (p *parser) parseGoBlock() exprAST {
	return &goBlockAST{code: p.next().val}
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
