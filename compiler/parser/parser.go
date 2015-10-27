package parser

import (
	"fmt"
	"github.com/defiant00/char/compiler/ast"
	"github.com/defiant00/char/compiler/lexer"
	"github.com/defiant00/char/compiler/token"
	"io/ioutil"
)

type parser struct {
	fileName  string        // file name being parsed
	pos       int           // current position in the token slice
	tokens    []token.Token // all relevant program tokens
	fmtTokens []token.Token // all tokens, used to format
}

func Parse(file string, build, format, printTokens bool) ast.General {
	fmt.Println("Parsing file", file)

	dat, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	input := string(dat)
	fmt.Println("Data loaded...")

	l := lexer.Lex(input)
	p := parser{fileName: file}
	var t token.Token

	if printTokens {
		fmt.Println("\nTokens")
	}

	// Read all tokens into a slice.
	for {
		t = l.NextToken()
		if format {
			p.fmtTokens = append(p.fmtTokens, t)
		}
		if build && t.Type != token.SLCOMMENT {
			p.tokens = append(p.tokens, t)
		}
		if printTokens {
			fmt.Print(" ", t)
		}
		if t.Type == token.ERROR || t.Type == token.EOF {
			break
		}
	}
	if t.Type == token.ERROR {
		return p.errorStmt(false, "\n\n%v\n", t)
	}
	return p.parseFile()
}

func (p *parser) errorStmt(toNextLine bool, format string, args ...interface{}) ast.Statement {
	p.toNextLine(toNextLine)
	return &ast.Error{Val: fmt.Sprintf(format, args...)}
}

func (p *parser) errorExpr(toNextLine bool, format string, args ...interface{}) ast.Expression {
	p.toNextLine(toNextLine)
	return &ast.Error{Val: fmt.Sprintf(format, args...)}
}

func (p *parser) toNextLine(toNextLine bool) {
	if !toNextLine {
		return
	}

	t := p.peek().Type
	for t != token.EOL && t != token.EOF {
		p.next()
		t = p.peek().Type
	}
	if t == token.EOL {
		p.next()
		for p.peek().Type == token.DEDENT {
			p.next()
		}
	}
}

func (p *parser) peek() token.Token {
	return p.tokens[p.pos]
}

func (p *parser) next() token.Token {
	t := p.peek()
	p.pos++
	return t
}

func (p *parser) backup(count int) {
	p.pos -= count
}

func (p *parser) accept(types ...token.Type) (bool, []token.Token) {
	start := p.pos
	var tokens []token.Token
	var typ token.Type
	for len(types) > 0 {
		typ, types = types[0], types[1:]
		cur := p.next()
		tokens = append(tokens, cur)
		if cur.Type != typ {
			p.pos = start
			return false, tokens
		}
	}
	return true, tokens
}

func (p *parser) parseFile() ast.General {
	f := &ast.File{Name: p.fileName}
	for p.pos < len(p.tokens) {
		switch p.peek().Type {
		case token.EOF:
			p.next()
		case token.IDENTIFIER:
			f.AddStmt(p.parseTopLevelIdent())
		case token.FUNCTION:
			f.AddStmt(p.parseTypeRedirect())
		case token.USE:
			f.AddStmt(p.parseUse())
		default:
			f.AddStmt(p.errorStmt(true, "Unknown token %v", p.peek()))
		}
	}
	return f
}

func (p *parser) parseTopLevelIdent() ast.Statement {
	p.next() // eat identifier
	t := p.peek()
	p.backup(1)
	if t.Type == token.DOT || t.Type == token.AS {
		return p.parseTypeRedirect()
	}
	return p.parseClass()
}

func (p *parser) parseTypeRedirect() ast.Statement {
	t := &ast.TypeRedirect{Type: p.parseType()}

	succ, toks := p.accept(token.AS)
	if !succ {
		return p.errorStmt(true, "Unknown token in type redirect: %v", toks[len(toks)-1])
	}

	succ, toks = p.accept(token.IDENTIFIER, token.EOL)
	if !succ {
		return p.errorStmt(true, "Unknown token in type redirect: %v", toks[len(toks)-1])
	}
	t.Name = toks[0].Val
	return t
}

func (p *parser) parseType() ast.Statement {
	if p.peek().Type == token.IDENTIFIER {
		return p.parseIdentifier()
	}
	return p.parseAnonFuncType()
}

func (p *parser) parseAnonFuncType() ast.Statement {
	a := &ast.AnonFuncType{}

	// func(types)
	succ, toks := p.accept(token.FUNCTION, token.LEFTPAREN)
	if !succ {
		return p.errorStmt(true, "Unknown token in anonymous function: %v", toks[len(toks)-1])
	}
	for p.peek().Type != token.RIGHTPAREN {
		a.AddParam(p.parseType())
		switch p.peek().Type {
		case token.COMMA:
			p.next() // eat ,
		case token.RIGHTPAREN:
		default:
			return p.errorStmt(true, "Unknown token in anonymous function: %v", p.peek())
		}
	}
	p.next() // eat )

	// return value(s)
	switch t := p.peek().Type; {
	case t.IsType():
		a.AddReturn(p.parseType())
	case t == token.LEFTPAREN:
		p.next() // eat (
		for p.peek().Type != token.RIGHTPAREN {
			a.AddReturn(p.parseType())
			switch p.peek().Type {
			case token.COMMA:
				p.next() // eat ,
			case token.RIGHTPAREN:
			default:
				return p.errorStmt(true, "Unknown token in anonymous function: %v", p.peek())
			}
		}
		p.next() // eat )
	}

	return a
}

func (p *parser) parseClass() ast.Statement {
	c := &ast.Class{Name: p.next().Val}

	succ, toks := p.accept(token.LEFTCARET)
	if succ {
		succ, toks = p.accept(token.IDENTIFIER)
		if !succ {
			return p.errorStmt(true, "Unknown token in class %v type declaration: %v", c.Name, toks[len(toks)-1])
		}
		c.AddTypeParam(toks[0].Val)
		for {
			succ, toks = p.accept(token.COMMA, token.IDENTIFIER)
			if !succ {
				break
			}
			c.AddTypeParam(toks[1].Val)
		}
		succ, toks = p.accept(token.RIGHTCARET)
		if !succ {
			return p.errorStmt(true, "Unknown token in class %v type declaration: %v", c.Name, toks[len(toks)-1])
		}
	}

	succ, toks = p.accept(token.WITH)
	if succ {
		if p.peek().Type != token.IDENTIFIER {
			return p.errorStmt(true, "Unknown token in class %v with declaration: %v", c.Name, p.peek())
		}
		c.AddWith(p.parseIdentifier())

		for {
			succ, toks = p.accept(token.COMMA)
			if !succ {
				break
			}
			if p.peek().Type != token.IDENTIFIER {
				return p.errorStmt(true, "Unknown token in class %v with declaration: %v", c.Name, p.peek())
			}
			c.AddWith(p.parseIdentifier())
		}
	}

	succ, toks = p.accept(token.EOL, token.INDENT)
	if !succ {
		return p.errorStmt(true, "Unknown token in class %v declaration: %v", c.Name, toks[len(toks)-1])
	}

	for p.peek().Type != token.DEDENT && p.peek().Type != token.EOF {
		c.AddStmt(p.parseClassStmt())
	}

	succ, toks = p.accept(token.DEDENT)
	if !succ {
		c.AddStmt(p.errorStmt(true, "Unknown token in class %v declaration: %v", c.Name, toks[len(toks)-1]))
	}

	return c
}

func (p *parser) parseClassStmt() ast.Statement {
	switch p.peek().Type {
	case token.DOT:
		return p.parseClassStmtIdent(true)
	case token.IDENTIFIER:
		return p.parseClassStmtIdent(false)
	case token.IOTA:
		return p.parseIotaStmt()
	}
	return p.errorStmt(true, "Unknown token in class statement: %v", p.peek())
}

func (p *parser) parseClassStmtIdent(dotted bool) ast.Statement {
	if dotted {
		p.next() // eat .
	}
	succ, toks := p.accept(token.IDENTIFIER)
	if !succ {
		return p.errorStmt(true, "Unknown token in class statement: %v", toks[len(toks)-1])
	}

	name := toks[0].Val
	if p.peek().Type == token.LEFTPAREN {
		return p.errorStmt(true, "Function parsing not yet implemented!")
	}

	var typ ast.Statement
	if p.peek().Type.IsType() {
		typ = p.parseType()
	}

	var val ast.Expression
	if p.peek().Type == token.ASSIGN {
		p.next() // eat =
		val = p.parseExpr()
	}

	succ, toks = p.accept(token.EOL)
	if !succ {
		return p.errorStmt(true, "Unknown token in class statement: %v", toks[len(toks)-1])
	}

	return &ast.Property{Static: !dotted, Name: name, Type: typ, Val: val}
}

func (p *parser) parsePrimaryExpr() ast.Expression {
	switch p.peek().Type {
	//case token.LEFTPAREN:
	//return p.parseParenExpr()
	//case token.IDENTIFIER:
	//return p.parseIdentExpr()
	case token.IOTA:
		return p.parseIotaExpr()
	case token.STRING:
		return p.parseStringExpr()
	case token.NUMBER:
		return p.parseNumberExpr()
	case token.CHAR:
		return p.parseCharExpr()
	case token.TRUE, token.FALSE:
		return p.parseBoolExpr()
	}
	return p.errorExpr(true, "Token is not an expression: %v", p.peek())
}

func (p *parser) parseBoolExpr() ast.Expression {
	t := p.next()
	return &ast.BoolExpr{Val: t.Type == token.TRUE}
}

func (p *parser) parseCharExpr() ast.Expression {
	return &ast.CharExpr{Val: p.next().Val}
}

func (p *parser) parseNumberExpr() ast.Expression {
	return &ast.NumberExpr{Val: p.next().Val}
}

func (p *parser) parseIotaExpr() ast.Expression {
	p.next()
	return &ast.IotaExpr{}
}

func (p *parser) parseStringExpr() ast.Expression {
	return &ast.StringExpr{Val: p.next().Val}
}

func (p *parser) parseExpr() ast.Expression {
	lhs := p.parsePrimaryExpr()
	switch lhs.(type) {
	case *ast.Error:
		return lhs
	}

	return p.parseBinopRHS(0, lhs)
}

func (p *parser) parseBinopRHS(exprPrec int, lhs ast.Expression) ast.Expression {
	for {
		tokPrec := p.peek().Precedence()

		// If this is a binary operator that binds as tightly as the
		// current one, consume it. Otherwise we're done.
		if tokPrec < exprPrec {
			return lhs
		}

		op := p.next()

		rhs := p.parsePrimaryExpr()
		switch rhs.(type) {
		case *ast.Error:
			return rhs // An error, so rhs should hold the error message
		}

		// If binop binds less tightly with RHS than the operator after RHS,
		// let the pending op take RHS as its LHS.
		nextPrec := p.peek().Precedence()
		if tokPrec < nextPrec {
			rhs = p.parseBinopRHS(tokPrec+1, rhs)
			switch rhs.(type) {
			case *ast.Error:
				return rhs // An error, so rhs should hold the error message
			}
		}

		// Merge lhs/rhs
		lhs = &ast.BinaryExpr{Op: op.Type, Left: lhs, Right: rhs}
	}
}

func (p *parser) parseIotaStmt() ast.Statement {
	succ, toks := p.accept(token.IOTA, token.EOL)
	if succ {
		return &ast.IotaStmt{}
	}
	return p.errorStmt(true, "Unknown token in iota reset: %v", toks[len(toks)-1])
}

func (p *parser) parseIdentifier() *ast.Identifier {
	i := &ast.Identifier{}
	i.AddIdent(p.next().Val)
	for {
		succ, toks := p.accept(token.DOT, token.IDENTIFIER)
		if !succ {
			break
		}
		i.AddIdent(toks[1].Val)
	}
	return i
}

func (p *parser) parseUse() ast.Statement {
	p.next() // eat token.USE
	u := &ast.Use{}
	succ, _ := p.accept(token.EOL, token.INDENT)
	if succ {
		err, pack, alias, errTok := p.parseUsePackage()
		for !err {
			u.AddPackage(pack, alias)
			err, pack, alias, errTok = p.parseUsePackage()
		}
		succ, _ = p.accept(token.DEDENT)
		if succ {
			return u
		}
		return p.errorStmt(true, "Unknown token found when parsing Use: %v", errTok)
	}

	err, pack, alias, errTok := p.parseUsePackage()
	if err {
		return p.errorStmt(true, "Unknown token found when parsing Use: %v", errTok)
	}
	u.AddPackage(pack, alias)

	return u
}

func (p *parser) parseUsePackage() (bool, string, string, token.Token) {
	succ, toks := p.accept(token.STRING, token.EOL)
	if succ {
		return false, toks[0].Val, "", toks[0]
	}
	succ, toks = p.accept(token.STRING, token.AS, token.IDENTIFIER, token.EOL)
	if succ {
		return false, toks[0].Val, toks[2].Val, toks[0]
	}
	return true, "", "", toks[len(toks)-1]
}
