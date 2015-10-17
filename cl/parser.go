package cl

import (
	"fmt"
)

type parser struct {
	tokens []token // All tokens
	pos    int     // Current position in tokens
	run    bool    // Whether the parser is running
}

func (p *parser) next() token {
	n := p.peek()
	p.pos++
	return n
}

func (p *parser) peek() token {
	return p.tokens[p.pos]
}

func (p *parser) errorf(format string, args ...interface{}) genAST {
	p.run = false
	return &errorAST{error: fmt.Sprintf(format, args...)}
}

func (p *parser) acceptTokens(types ...tType) (bool, []token) {
	start := p.pos
	var tokens []token
	var typ tType
loop:
	for len(types) > 0 {
		typ = types[0]
		cur := p.next()
		// Eat any comments
		for cur.typ == tSLComment {
			cur = p.next()
		}
		switch typ {
		case tEOL:
			if cur.typ == tEOL {
				types = types[1:]
				tokens = append(tokens, cur)
				// Eat any extra EOLs or comments.
				for p.peek().typ == tEOL || p.peek().typ == tSLComment {
					p.next()
				}
			} else {
				tokens = append(tokens, cur) // Append current token on error for better errors.
				break loop
			}
		case cur.typ:
			types = types[1:]
			tokens = append(tokens, cur)
		default:
			tokens = append(tokens, cur) // Append current token on error for better errors.
			break loop
		}
	}
	if len(types) == 0 {
		return true, tokens
	}

	p.pos = start
	return false, tokens
}

func (p *parser) parseFile() genAST {
	f := &fileAST{}
	for p.run {
		switch p.peek().typ {
		case tEOF:
			p.run = false
		case tEOL, tSLComment:
			p.next()
		case tGoBlock:
			f.addStmt(p.parseGoBlock())
		case tUse:
			f.addStmt(p.parseUse())
		case tIdentifier:
			f.addStmt(p.parseClass())
		default:
			f.addStmt(p.errorf("Unknown token %v", p.peek()))
		}
	}
	return f
}

func (p *parser) parseGoBlock() genAST {
	return &goBlockAST{block: p.next().val}
}

func (p *parser) parseUse() genAST {
	u := &useAST{}
	succ, toks := p.acceptTokens(tUse, tEOL, tIndent)
	if !succ {
		return p.errorf("Invalid token in a use: %v", toks[len(toks)-1])
	}
	// Eat the use statements.
	succ, toks = p.acceptTokens(tString, tEOL)
	for succ {
		u.addPkg(toks[0].val)
		succ, toks = p.acceptTokens(tString, tEOL)
	}
	// Eat the dedent.
	succ, toks = p.acceptTokens(tDedent)
	if !succ {
		return p.errorf("Invalid token in a use: %v", toks[len(toks)-1])
	}
	return u
}

func (p *parser) parseClass() genAST {
	c := &classAST{}
	succ, toks := p.acceptTokens(tIdentifier, tEOL, tIndent)
	if !succ {
		return p.errorf("Invalid token in a class definition: %v", toks[len(toks)-1])
	}
	c.name = toks[0].val

loop:
	for p.run {
		switch p.peek().typ {
		case tDedent:
			p.next()
			break loop
		case tIdentifier:
			t := p.next()
			switch p.peek().typ {
			case tLeftParen:
				c.addStmt(p.parseFunction(t.val, true))
			case tIdentifier:
				c.addStmt(&classVarAST{name: t.val, typ: p.parseExpr()})
				succ, _ = p.acceptTokens(tEOL)
				if !succ {
					c.addStmt(p.errorf("Invalid token in class %v: %v", c.name, p.peek()))
				}
			default:
				c.addStmt(p.errorf("Invalid token in class %v: %v", c.name, p.peek()))
			}
		case tDot:
			succ, toks = p.acceptTokens(tDot, tIdentifier, tLeftParen)
			if !succ {
				c.addStmt(p.errorf("Invalid token in class %v: %v", c.name, toks[len(toks)-1]))
				break loop
			}
			c.addStmt(p.parseFunction(toks[1].val, false))
		case tConst:
			c.addStmt(p.parseConsts())
		default:
			c.addStmt(p.errorf("Invalid token in class %v: %v", c.name, p.peek()))
		}
	}
	return c
}

func (p *parser) parseConsts() genAST {
	succ, toks := p.acceptTokens(tConst, tEOL, tIndent)
	if succ {
		c := &constAST{}
	loop:
		for p.run {
			switch p.peek().typ {
			case tDedent:
				p.next()
				break loop
			case tEOL, tSLComment:
				p.next()
			default:
				c.addDef(p.parseExpr())
			}
		}
		return c
	} else {
		return p.errorf("Invalid token in constant: %v", toks[len(toks)-1])
	}
}

func (p *parser) parseFunction(name string, static bool) genAST {
	p.next() // Eat the '('
	f := &funcAST{name: name, static: static}

	// TODO - Parameters

	succ, toks := p.acceptTokens(tRightParen, tEOL, tIndent)
	if !succ {
		return p.errorf("Invalid token in function definition: %v", toks[len(toks)-1])
	}

	for p.peek().typ != tDedent && p.run {
		switch p.peek().typ {
		case tVar:
			f.addExpr(p.parseVarDeclaration())
		case tEOL, tSLComment:
			p.next()
		case tGoBlock:
			f.addExpr(p.parseGoBlock())
		default:
			f.addExpr(p.parseExpr())
		}
	}
	// Eat the dedent.
	p.next()

	return f
}

func (p *parser) parsePrimaryExpr() exprAST {
	switch p.peek().typ {
	case tIdentifier:
		return p.parseIdentExpr()
	case tString:
		return p.parseStringExpr()
	case tNumber:
		return p.parseNumberExpr()
	default:
		return p.errorf("Token is not an expression: %v", p.peek())
	}
}

func (p *parser) parseExpr() exprAST {
	lhs := p.parsePrimaryExpr()
	if !p.run {
		return lhs
	}

	return p.parseBinopRHS(0, lhs)
}

func (p *parser) parseBinopRHS(exprPrec int, lhs exprAST) exprAST {
	var rhs exprAST
	for p.run {
		tokPrec := p.peek().precedence()

		// If this is a binary operator that binds as tightly as the
		// current one, consume it. Otherwise we're done.
		if tokPrec < exprPrec {
			return lhs
		}

		op := p.next()

		rhs = p.parsePrimaryExpr()
		if !p.run {
			return rhs // An error, so rhs should hold the error message
		}

		// If binop binds less tightly with RHS than the operator after RHS,
		// let the pending op take RHS as its LHS.
		nextPrec := p.peek().precedence()
		if tokPrec < nextPrec {
			rhs = p.parseBinopRHS(tokPrec+1, rhs)
			if !p.run {
				return rhs // An error, rhs should hold the error message
			}
		}

		// Merge lhs/rhs
		lhs = &binaryExprAST{op: op.typ, left: lhs, right: rhs}
	}
	return rhs // An error, rhs should hold the error message
}

func (p *parser) parseVarDeclaration() exprAST {
	// Eat 'var'
	p.next()
	return &varDeclareAST{initial: p.parseExpr()}
}

func (p *parser) parseIdentExpr() exprAST {
	return &identExprAST{name: p.next().val}
}

func (p *parser) parseStringExpr() exprAST {
	return &stringExprAST{val: p.next().val}
}

func (p *parser) parseNumberExpr() exprAST {
	return &numberExprAST{val: p.next().val}
}

func parse(input string) genAST {
	p := &parser{run: true}
	l := lex(input)
	var t token

	// Read all tokens into a slice.
	for {
		t = l.nextToken()
		p.tokens = append(p.tokens, t)
		//fmt.Print(" ", t)
		if t.typ == tError || t.typ == tEOF {
			break
		}
	}
	if t.typ == tError {
		return p.errorf("Error token encountered: %v", t)
	}
	return p.parseFile()
}
