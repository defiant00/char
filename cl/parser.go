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

func (p *parser) peek() token {
	return p.tokens[p.pos]
}

func (p *parser) errorf(format string, args ...interface{}) exprAST {
	return errorAST{error: fmt.Sprintf(format, args...)}
}

func (p *parser) parseProgram() exprAST {
	e := &programAST{}
	// Build the AST
loop:
	for {
		switch p.peek().typ {
		case tError, tEOF:
			break loop
		case tGoBlock:
			e.add(p.parseGoBlock())
		case tUse:
			e.add(p.parseUse())
		case tIdentifier:
			e.add(p.parseClass())
		case tEOL:
			p.next()
		default:
			e.add(p.errorf("Unknown token %v", p.peek()))
			break loop
		}
	}
	return e
}

func (p *parser) parseUse() exprAST {
	u := &useAST{}
	p.next() // Consume the use token
	t := p.next()
	if t.typ != tEOL {
		return p.errorf("Not an EOL: %v", t)
	}
	for p.peek().typ == tEOL {
		p.next()
	}
	t = p.next()
	if t.typ != tIndent {
		return p.errorf("Not an indent: %v", t)
	}
	for {
		t = p.next()
		if t.typ == tDedent {
			break
		}
		if t.typ != tString {
			return p.errorf("Not a package name: %v", t)
		}
		u.packages = append(u.packages, t.val)
		t = p.next()
		if t.typ != tEOL {
			return p.errorf("Not an EOL: %v", t)
		}
		for p.peek().typ == tEOL {
			p.next()
		}
	}
	return u
}

func (p *parser) parseClass() exprAST {
	c := &classAST{}
	t := p.next() // Consume the identifier
	c.name = t.val
	t = p.next()
	if t.typ != tEOL {
		return p.errorf("Not an EOL: %v", t)
	}
	for p.peek().typ == tEOL {
		p.next()
	}
	t = p.next()
	if t.typ != tIndent {
		return p.errorf("Not an indent: %v", t)
	}
loop:
	for {
		switch p.peek().typ {
		case tEOL:
			p.next()
		case tIdentifier:
			c.items = append(c.items, p.parseClassIdentifier(false))
		case tDot:
			p.next() // Consume the dot
			if p.peek().typ != tIdentifier {
				c.items = append(c.items, p.errorf("Not an identifier: %v", p.peek()))
				return c
			}
			c.items = append(c.items, p.parseClassIdentifier(true))
		case tDedent:
			p.next()
			break loop
		default:
			c.items = append(c.items, p.errorf("Expecting a method or variable, found %v", p.peek()))
			return c
		}
	}
	return c
}

func (p *parser) parseClassIdentifier(dotted bool) exprAST {
	t := p.next()
	if p.peek().typ == tLeftParen {
		p.next() // Consume (
		if p.peek().typ != tRightParen {
			return p.errorf("Not right paren: %v", p.peek())
		}
		p.next() // Consume )
		t2 := p.next()
		if t2.typ != tEOL {
			return p.errorf("Not an EOL: %v", t2)
		}
		for p.peek().typ == tEOL {
			p.next()
		}
		t2 = p.next()
		if t2.typ != tIndent {
			return p.errorf("Not an indent: %v", t2)
		}
		return funcDefAST{name: t.val, static: !dotted}
	}
	typ := p.next()
	if typ.typ != tIdentifier {
		return p.errorf("Not an identifier: %v", typ)
	}
	return varDefAST{name: t.val, typ: typ.val, static: !dotted}
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
