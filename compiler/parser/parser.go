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
		if build && t.Type != token.COMMENT {
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
		case token.MIXIN:
			f.AddStmt(p.parseMixin())
		default:
			f.AddStmt(p.errorStmt(true, "Invalid token %v", p.peek()))
		}
	}
	return f
}

func (p *parser) parseTopLevelIdent() ast.Statement {
	if p.isTypeRedirect() {
		return p.parseTypeRedirect()
	}
	return p.parseClass(false)
}

// isTypeRedirect returns whether a line of tokens is a type redirect.
// It reads through tokens until it encounters an EOL. During that time,
// if it encounters an AS it returns true, otherwise false.
func (p *parser) isTypeRedirect() bool {
	count := 0
	for p.peek().Type != token.EOL {
		if p.peek().Type == token.AS {
			p.backup(count)
			return true
		}
		p.next()
		count++
	}
	p.backup(count)
	return false
}

func (p *parser) parseMixin() ast.Statement {
	p.next() // eat mix
	return p.parseClass(true)
}

func (p *parser) parseTypeRedirect() ast.Statement {
	t := &ast.TypeRedirect{Type: p.parseType()}

	if succ, toks := p.accept(token.AS); !succ {
		return p.errorStmt(true, "Invalid token in type redirect: %v", toks[len(toks)-1])
	}

	succ, toks := p.accept(token.IDENTIFIER, token.EOL)
	if !succ {
		return p.errorStmt(true, "Invalid token in type redirect: %v", toks[len(toks)-1])
	}
	t.Name = toks[0].Val
	return t
}

func (p *parser) parseType() ast.Statement {
	if p.peek().Type == token.IDENTIFIER {
		return p.parseTypeIdent()
	}
	return p.parseAnonFuncType()
}

func (p *parser) parseAnonFuncType() ast.Statement {
	a := &ast.AnonFuncType{}

	// func(types)
	if succ, toks := p.accept(token.FUNCTION, token.LEFTPAREN); !succ {
		return p.errorStmt(true, "Invalid token in anonymous function: %v", toks[len(toks)-1])
	}
	for p.peek().Type != token.RIGHTPAREN {
		a.AddParam(p.parseType())
		switch p.peek().Type {
		case token.COMMA:
			p.next() // eat ,
		case token.RIGHTPAREN:
		default:
			return p.errorStmt(true, "Invalid token in anonymous function: %v", p.peek())
		}
	}
	p.next() // eat )

	// return value(s)
	rvs := p.parseReturnValues()
	for _, rv := range rvs {
		a.AddReturn(rv)
	}

	return a
}

func (p *parser) parseClass(mixin bool) ast.Statement {
	succ, toks := p.accept(token.IDENTIFIER)
	if !succ {
		return p.errorStmt(true, "Invalid token in class declaration: %v", toks[len(toks)-1])
	}
	c := &ast.Class{Mixin: mixin, Name: toks[0].Val}

	if succ, _ := p.accept(token.LEFTCARET); succ {
		succ, toks := p.accept(token.IDENTIFIER)
		if !succ {
			return p.errorStmt(true, "Invalid token in class %v type declaration: %v", c.Name, toks[len(toks)-1])
		}
		c.AddTypeParam(toks[0].Val)
		for {
			succ, toks = p.accept(token.COMMA, token.IDENTIFIER)
			if !succ {
				break
			}
			c.AddTypeParam(toks[1].Val)
		}
		if succ, _ = p.accept(token.RIGHTCARET); !succ {
			return p.errorStmt(true, "Invalid token in class %v type declaration: %v", c.Name, toks[len(toks)-1])
		}
	}

	if succ, _ := p.accept(token.WITH); succ {
		if p.peek().Type != token.IDENTIFIER {
			return p.errorStmt(true, "Invalid token in class %v with declaration: %v", c.Name, p.peek())
		}
		c.AddWith(p.parseTypeIdent())

		for {
			if succ, _ = p.accept(token.COMMA); !succ {
				break
			}
			if p.peek().Type != token.IDENTIFIER {
				return p.errorStmt(true, "Invalid token in class %v with declaration: %v", c.Name, p.peek())
			}
			c.AddWith(p.parseTypeIdent())
		}
	}

	if succ, toks := p.accept(token.EOL, token.INDENT); !succ {
		return p.errorStmt(true, "Invalid token in class %v declaration: %v", c.Name, toks[len(toks)-1])
	}

	for p.peek().Type != token.DEDENT && p.peek().Type != token.EOF {
		c.AddStmt(p.parseClassStmt())
	}

	if succ, toks := p.accept(token.DEDENT); !succ {
		c.AddStmt(p.errorStmt(true, "Invalid token in class %v declaration: %v", c.Name, toks[len(toks)-1]))
	}

	return c
}

func (p *parser) parseClassStmt() ast.Statement {
	switch p.peek().Type {
	case token.DOT, token.IDENTIFIER:
		return p.parseClassStmtIdent()
	case token.IOTA:
		return p.parseIotaStmt()
	}
	return p.errorStmt(true, "Invalid token in class statement: %v", p.peek())
}

func (p *parser) parseFuncStmt() ast.Statement {
	switch p.peek().Type {
	case token.VAR:
		return p.parseFuncStmtVar()
	default:
		return p.parseExprStmt()
	}
}

func (p *parser) parseExprStmt() ast.Statement {
	ex := p.parseExpr()
	switch ex.(type) {
	case *ast.Error:
	default:
		if succ, toks := p.accept(token.EOL); !succ {
			return p.errorStmt(true, "Invalid token in expression statement: %v", toks[len(toks)-1])
		}
	}
	return &ast.ExprStmt{Expr: ex}
}

func (p *parser) parseFuncStmtVar() ast.Statement {
	p.next() // eat var
	vs := &ast.VarSet{}

	vsl := p.parseFuncStmtVarLine()
	switch vsl.(type) {
	case *ast.Error:
		return vsl
	}
	vs.AddLine(vsl.(*ast.VarSetLine))

	if succ, _ := p.accept(token.INDENT); succ {
		for p.peek().Type != token.DEDENT && p.peek().Type != token.EOF {
			vsl = p.parseFuncStmtVarLine()
			switch vsl.(type) {
			case *ast.Error:
				return vsl
			}
			vs.AddLine(vsl.(*ast.VarSetLine))
		}

		if succ, toks := p.accept(token.DEDENT); !succ {
			return p.errorStmt(true, "Invalid token in var statement: %v", toks[len(toks)-1])
		}
	}

	return vs
}

func (p *parser) parseFuncStmtVarLine() ast.Statement {
	v := &ast.VarSetLine{}
	for {
		succ, toks := p.accept(token.IDENTIFIER)
		if !succ {
			return p.errorStmt(true, "Invalid token in var statement: %v", toks[len(toks)-1])
		}
		name := toks[0].Val

		var typ ast.Statement
		if p.peek().Type.IsType() {
			typ = p.parseType()
		}

		v.AddVar(name, typ)
		if succ, _ = p.accept(token.COMMA); !succ {
			break
		}
	}

	if succ, _ := p.accept(token.ASSIGN); succ {
		v.Vals = p.parseExpr()
	}

	if succ, toks := p.accept(token.EOL); !succ {
		return p.errorStmt(true, "Invalid token in var statement: %v", toks[len(toks)-1])
	}
	return v
}

func (p *parser) parseClassStmtIdent() ast.Statement {
	ps := &ast.PropertySet{}

	for {
		dotted, _ := p.accept(token.DOT)
		succ, toks := p.accept(token.IDENTIFIER)
		if !succ {
			return p.errorStmt(true, "Invalid token in class statement: %v", toks[len(toks)-1])
		}
		name := toks[0].Val

		var typ ast.Statement
		switch t := p.peek().Type; {
		case t == token.LEFTPAREN:
			return p.parseFunctionDef(dotted, name)
		case t.IsType():
			typ = p.parseType()
		}

		ps.AddProp(!dotted, name, typ)
		if succ, _ = p.accept(token.COMMA); !succ {
			break
		}
	}

	if succ, _ := p.accept(token.ASSIGN); succ {
		ps.Vals = p.parseExpr()
	}

	if succ, toks := p.accept(token.EOL); !succ {
		return p.errorStmt(true, "Invalid token in class statement: %v", toks[len(toks)-1])
	}

	return ps
}

func (p *parser) parseReturnValues() []ast.Statement {
	rvs := make([]ast.Statement, 0, 1)

	switch t := p.peek().Type; {
	case t.IsType():
		rvs = append(rvs, p.parseType())
	case t == token.LEFTPAREN:
		p.next() // eat (
		for p.peek().Type != token.RIGHTPAREN {
			rvs = append(rvs, p.parseType())
			switch p.peek().Type {
			case token.COMMA:
				p.next() // eat ,
			case token.RIGHTPAREN:
			default:
				return append(rvs, p.errorStmt(true, "Invalid token in return types: %v", p.peek()))
			}
		}
		p.next() // eat )
	}

	return rvs
}

// parseFunctionDef parses a function definition with the optional dot and name
// already consumed.
func (p *parser) parseFunctionDef(dotted bool, name string) ast.Statement {
	p.next() // eat (
	f := &ast.FuncDefStmt{Static: !dotted, Name: name}
	for p.peek().Type != token.RIGHTPAREN {
		succ, toks := p.accept(token.IDENTIFIER)
		if !succ {
			return p.errorStmt(true, "Invalid token in function %v definition: %v", name, p.peek())
		}
		name := toks[0].Val
		var typ ast.Statement
		if p.peek().Type.IsType() {
			typ = p.parseType()
		}
		f.AddParam(name, typ)
		switch p.peek().Type {
		case token.COMMA:
			p.next() // eat ,
		case token.RIGHTPAREN:
		default:
			return p.errorStmt(true, "Invalid token in function %v definition: %v", name, p.peek())
		}
	}
	if succ, toks := p.accept(token.RIGHTPAREN); !succ {
		return p.errorStmt(true, "Invalid token in function %v definition: %v", name, toks[len(toks)-1])
	}

	// return value(s)
	rvs := p.parseReturnValues()
	for _, rv := range rvs {
		f.AddReturn(rv)
	}

	if succ, toks := p.accept(token.EOL, token.INDENT); !succ {
		return p.errorStmt(true, "Invalid token in function %v definition: %v", name, toks[len(toks)-1])
	}

	for p.peek().Type != token.DEDENT && p.peek().Type != token.EOF {
		f.AddStmt(p.parseFuncStmt())
	}

	if succ, toks := p.accept(token.DEDENT); !succ {
		f.AddStmt(p.errorStmt(true, "Invalid token in function %v definition: %v", name, toks[len(toks)-1]))
	}

	return f
}

func (p *parser) parsePrimaryExpr() ast.Expression {
	switch p.peek().Type {
	case token.LEFTPAREN:
		return p.parseParenExpr()
	case token.IDENTIFIER:
		return p.parseIdentExpr()
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

func (p *parser) parseParenExpr() ast.Expression {
	p.next() // eat (
	expr := p.parseExpr()
	if succ, toks := p.accept(token.RIGHTPAREN); !succ {
		return p.errorExpr(true, "Invalid token in (): %v", toks[len(toks)-1])
	}
	return expr
}

func (p *parser) parseIdentExpr() ast.Expression {
	ips := make([]*ast.IdentPart, 0, 1)
	for {
		succ, toks := p.accept(token.IDENTIFIER)
		if !succ {
			return p.errorExpr(true, "Invalid token in identifier: %v", toks[len(toks)-1])
		}
		ip := &ast.IdentPart{Name: toks[0].Val}
		if succ, _ := p.accept(token.LEFTCARET); succ {
			resetPos := p.pos - 1 // store the position in case the caret isn't a generic
			for p.peek().Type.IsType() {
				ip.AddTypeParam(p.parseType())
				if succ, _ = p.accept(token.COMMA); !succ {
					break
				}
			}
			if succ, _ := p.accept(token.RIGHTCARET); !succ {
				p.pos = resetPos     // no closing caret, so reset position
				ip.ResetTypeParams() // and reset type parameters
			}
		}
		ips = append(ips, ip)
		if succ, _ := p.accept(token.DOT); !succ {
			break
		}
	}
	if p.peek().Type == token.LEFTPAREN {
		p.next() // eat (
		fc := &ast.FuncCallExpr{Idents: ips}
		if p.peek().Type != token.RIGHTPAREN {
			fc.Params = p.parseExpr()
		}
		if succ, toks := p.accept(token.RIGHTPAREN); !succ {
			return p.errorExpr(true, "Invalid token in function call: %v", toks[len(toks)-1])
		}
		return fc
	}
	return &ast.IdentExpr{Idents: ips}
}

func (p *parser) parseBoolExpr() ast.Expression {
	return &ast.BoolExpr{Val: p.next().Type == token.TRUE}
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
	if succ, toks := p.accept(token.IOTA, token.EOL); !succ {
		return p.errorStmt(true, "Invalid token in iota reset: %v", toks[len(toks)-1])
	}
	return &ast.IotaStmt{}
}

func (p *parser) parseTypeIdent() ast.Statement {
	t := &ast.TypeIdent{}
	t.AddIdent(p.next().Val)
	for {
		succ, toks := p.accept(token.DOT, token.IDENTIFIER)
		if !succ {
			break
		}
		t.AddIdent(toks[1].Val)
	}
	if succ, _ := p.accept(token.LEFTCARET); succ {
		for p.peek().Type.IsType() {
			t.AddTypeParam(p.parseType())
			if succ, _ = p.accept(token.COMMA); !succ {
				break
			}
		}
		if succ, toks := p.accept(token.RIGHTCARET); !succ {
			return p.errorStmt(true, "Invalid token parsing type identifier: %v", toks[len(toks)-1])
		}
	}
	return t
}

func (p *parser) parseUse() ast.Statement {
	p.next() // eat token.USE
	u := &ast.Use{}

	err, pack, alias, errTok := p.parseUsePackage()
	if err {
		return p.errorStmt(true, "Invalid token found when parsing Use: %v", errTok)
	}
	u.AddPackage(pack, alias)

	if succ, _ := p.accept(token.INDENT); succ {
		err, pack, alias, errTok := p.parseUsePackage()
		for !err {
			u.AddPackage(pack, alias)
			err, pack, alias, errTok = p.parseUsePackage()
		}
		if succ, _ = p.accept(token.DEDENT); succ {
			return u
		}
		return p.errorStmt(true, "Invalid token found when parsing Use: %v", errTok)
	}

	return u
}

func (p *parser) parseUsePackage() (bool, string, string, token.Token) {
	if succ, toks := p.accept(token.STRING, token.EOL); succ {
		return false, toks[0].Val, "", toks[0]
	}
	succ, toks := p.accept(token.STRING, token.AS, token.IDENTIFIER, token.EOL)
	if succ {
		return false, toks[0].Val, toks[2].Val, toks[0]
	}
	return true, "", "", toks[len(toks)-1]
}
