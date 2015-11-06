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

func Parse(file string, build, format, printTokens bool) (ast.General, bool) {
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
	return p.parseFile(), false
}

func (p *parser) errorStmt(toNextLine bool, format string, args ...interface{}) (ast.Statement, bool) {
	p.toNextLine(toNextLine)
	return &ast.Error{Val: fmt.Sprintf(format, args...)}, true
}

func (p *parser) errorExpr(toNextLine bool, format string, args ...interface{}) (ast.Expression, bool) {
	p.toNextLine(toNextLine)
	return &ast.Error{Val: fmt.Sprintf(format, args...)}, true
}

func (p *parser) toNextLine(toNextLine bool) {
	if !toNextLine {
		return
	}

	t := p.peek().Type
	for ; t != token.EOL && t != token.EOF; t = p.peek().Type {
		p.next()
	}
	if t == token.EOL {
		p.next()
		for {
			if succ, _ := p.accept(token.DEDENT, token.EOL); !succ {
				return
			}
		}
	}
}

func (p *parser) tokensAvailable() int {
	return len(p.tokens) - p.pos
}

func (p *parser) peek() token.Token {
	return p.tokens[p.pos]
}

func (p *parser) next() token.Token {
	t := p.peek()
	p.pos++
	return t
}

// peekCombo returns the next token, or combines >>
// into rshift.
func (p *parser) peekCombo() token.Token {
	t := p.next()
	if p.tokensAvailable() > 0 {
		t2 := p.peek()
		if t.Type == token.RIGHT_CARET && t2.Type == token.RIGHT_CARET {
			p.backup(1)
			return token.Token{Type: token.RSHIFT, Pos: t.Pos}
		}
	}
	p.backup(1)
	return t
}

// nextCombo consumes and returns the next token, or
// combines >> into rshift.
func (p *parser) nextCombo() token.Token {
	t := p.next()
	t2 := p.next()
	if t.Type == token.RIGHT_CARET && t2.Type == token.RIGHT_CARET {
		return token.Token{Type: token.RSHIFT, Pos: t.Pos}
	}
	p.backup(1)
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
			st, _ := p.parseTopLevelIdent()
			f.AddStmt(st)
		case token.FUNCTION:
			st, _ := p.parseTypeRedirect()
			f.AddStmt(st)
		case token.USE:
			st, _ := p.parseUse()
			f.AddStmt(st)
		case token.INTERFACE:
			st, _ := p.parseInterface()
			f.AddStmt(st)
		case token.MIXIN:
			st, _ := p.parseMixin()
			f.AddStmt(st)
		default:
			st, _ := p.errorStmt(true, "Invalid token %v", p.peek())
			f.AddStmt(st)
		}
	}
	return f
}

func (p *parser) parseInterface() (ast.Statement, bool) {
	succ, toks := p.accept(token.INTERFACE, token.IDENTIFIER)
	if !succ {
		return p.errorStmt(true, "Invalid token in interface: %v", toks[len(toks)-1])
	}

	intf := &ast.Interface{Name: toks[1].Val}

	if succ, _ := p.accept(token.WITH); succ {
		for {
			if p.peek().Type != token.IDENTIFIER {
				return p.errorStmt(true, "Invalid token in interface %v: %v", intf.Name, p.peek())
			}
			st, _ := p.parseTypeIdent()
			intf.AddWith(st)

			if succ, _ = p.accept(token.COMMA); !succ {
				break
			}
		}
	}

	if succ, toks = p.accept(token.EOL, token.INDENT); !succ {
		return p.errorStmt(true, "Invalid token in interface %v: %v", intf.Name, toks[len(toks)-1])
	}

	for p.peek().Type != token.DEDENT {
		// function_name(types)
		succ, toks = p.accept(token.IDENTIFIER, token.LEFT_PAREN)
		if !succ {
			return p.errorStmt(true, "Invalid token in interface %v: %v", intf.Name, toks[len(toks)-1])
		}
		fs := &ast.IntfFuncSig{Name: toks[0].Val}
		for p.peek().Type != token.RIGHT_PAREN {
			st, _ := p.parseType()
			fs.AddParam(st)
			switch p.peek().Type {
			case token.COMMA:
				p.next() // eat ,
			case token.RIGHT_PAREN:
			default:
				return p.errorStmt(true, "Invalid token in interface %v function signature %v: %v", intf.Name, fs.Name, p.peek())
			}
		}
		p.next() // eat )

		// return value(s)
		rvs, _ := p.parseReturnValues()
		for _, rv := range rvs {
			fs.AddReturn(rv)
		}

		if succ, _ = p.accept(token.EOL); !succ {
			return p.errorStmt(true, "Invalid token in interface %v function signature %v: %v", intf.Name, fs.Name, p.peek())
		}
		intf.AddFuncSig(fs)
	}

	if succ, toks = p.accept(token.DEDENT, token.EOL); !succ {
		return p.errorStmt(true, "Invalid token in interface %v: %v", intf.Name, toks[len(toks)-1])
	}

	return intf, false
}

func (p *parser) parseTopLevelIdent() (ast.Statement, bool) {
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

func (p *parser) parseMixin() (ast.Statement, bool) {
	p.next() // eat mix
	return p.parseClass(true)
}

func (p *parser) parseTypeRedirect() (ast.Statement, bool) {
	st, _ := p.parseType()
	t := &ast.TypeRedirect{Type: st}

	if succ, toks := p.accept(token.AS); !succ {
		return p.errorStmt(true, "Invalid token in type redirect: %v", toks[len(toks)-1])
	}

	succ, toks := p.accept(token.IDENTIFIER, token.EOL)
	if !succ {
		return p.errorStmt(true, "Invalid token in type redirect: %v", toks[len(toks)-1])
	}
	t.Name = toks[0].Val
	return t, false
}

func (p *parser) parseType() (ast.Statement, bool) {
	switch p.peek().Type {
	case token.IDENTIFIER:
		return p.parseTypeIdent()
	case token.ARRAY:
		return p.parseArrayType()
	case token.FUNCTION:
		return p.parseFuncSigType()
	default:
		return p.errorStmt(true, "Invalid token in type identifier: %v", p.peek())
	}
}

func (p *parser) parseArrayType() (ast.Statement, bool) {
	p.next() // eat []
	st, _ := p.parseType()
	return &ast.Array{Type: st}, false
}

func (p *parser) parseFuncSigType() (ast.Statement, bool) {
	f := &ast.FunctionSig{}

	// fn(types)
	if succ, toks := p.accept(token.FUNCTION, token.LEFT_PAREN); !succ {
		return p.errorStmt(true, "Invalid token in anonymous function: %v", toks[len(toks)-1])
	}
	for p.peek().Type != token.RIGHT_PAREN {
		st, _ := p.parseType()
		f.AddParam(st)
		switch p.peek().Type {
		case token.COMMA:
			p.next() // eat ,
		case token.RIGHT_PAREN:
		default:
			return p.errorStmt(true, "Invalid token in anonymous function: %v", p.peek())
		}
	}
	p.next() // eat )

	// return value(s)
	rvs, _ := p.parseReturnValues()
	for _, rv := range rvs {
		f.AddReturn(rv)
	}

	return f, false
}

func (p *parser) parseClass(mixin bool) (ast.Statement, bool) {
	succ, toks := p.accept(token.IDENTIFIER)
	if !succ {
		return p.errorStmt(true, "Invalid token in class declaration: %v", toks[len(toks)-1])
	}
	c := &ast.Class{Mixin: mixin, Name: toks[0].Val}

	if succ, _ := p.accept(token.LEFT_CARET); succ {
		for {
			succ, toks := p.accept(token.IDENTIFIER)
			if !succ {
				return p.errorStmt(true, "Invalid token in class %v type declaration: %v", c.Name, toks[len(toks)-1])
			}
			c.AddTypeParam(toks[0].Val)
			if succ, _ := p.accept(token.COMMA); !succ {
				break
			}
		}
		if succ, _ = p.accept(token.RIGHT_CARET); !succ {
			return p.errorStmt(true, "Invalid token in class %v type declaration: %v", c.Name, toks[len(toks)-1])
		}
	}

	if succ, _ := p.accept(token.WITH); succ {
		for {
			if p.peek().Type != token.IDENTIFIER {
				return p.errorStmt(true, "Invalid token in class %v with declaration: %v", c.Name, p.peek())
			}
			st, _ := p.parseTypeIdent()
			c.AddWith(st)

			if succ, _ = p.accept(token.COMMA); !succ {
				break
			}
		}
	}

	if succ, toks := p.accept(token.EOL, token.INDENT); !succ {
		return p.errorStmt(true, "Invalid token in class %v declaration: %v", c.Name, toks[len(toks)-1])
	}

	for p.peek().Type != token.DEDENT && p.peek().Type != token.EOF {
		st, _ := p.parseClassStmt()
		c.AddStmt(st)
	}

	if succ, toks := p.accept(token.DEDENT, token.EOL); !succ {
		st, _ := p.errorStmt(true, "Invalid token in class %v declaration: %v", c.Name, toks[len(toks)-1])
		c.AddStmt(st)
	}

	return c, false
}

func (p *parser) parseClassStmt() (ast.Statement, bool) {
	switch p.peek().Type {
	case token.DOT, token.IDENTIFIER:
		return p.parseClassStmtIdent()
	case token.IOTA:
		return p.parseIotaStmt()
	}
	return p.errorStmt(true, "Invalid token in class statement: %v", p.peek())
}

func (p *parser) parseFuncStmt() (ast.Statement, bool) {
	switch p.peek().Type {
	case token.VAR:
		return p.parseVarStmt(false)
	case token.RETURN:
		return p.parseReturnStmt()
	case token.DEFER:
		return p.parseDeferStmt()
	case token.IF:
		return p.parseIfStmt()
	case token.BREAK:
		return p.parseBreakStmt()
	case token.FOR, token.LOOP:
		return p.parseForOrLoop("")
	default:
		if succ, toks := p.accept(token.IDENTIFIER, token.COLON); succ {
			return p.parseForOrLoop(toks[0].Val)
		}
		return p.parseExprStmt(false)
	}
}

func (p *parser) parseBreakStmt() (ast.Statement, bool) {
	p.next() // eat break
	b := &ast.Break{}
	if succ, toks := p.accept(token.IDENTIFIER); succ {
		b.Label = toks[0].Val
	}
	if succ, toks := p.accept(token.EOL); !succ {
		return p.errorStmt(true, "Invalid token in break: %v", toks[len(toks)-1])
	}
	return b, false
}

func (p *parser) parseForOrLoop(label string) (ast.Statement, bool) {
	switch p.peek().Type {
	case token.FOR:
		return p.parseForStmt(label)
	case token.LOOP:
		return p.parseLoopStmt(label)
	default:
		return p.errorStmt(true, "Invalid token after label: %v", p.peek())
	}
}

func (p *parser) parseForStmt(label string) (ast.Statement, bool) {
	p.next() // eat for

	f := &ast.For{Label: label}
	for {
		succ, toks := p.accept(token.IDENTIFIER)
		if !succ {
			return p.errorStmt(true, "Invalid token in for: %v", toks[len(toks)-1])
		}
		f.AddVar(toks[0].Val)
		if succ, toks = p.accept(token.COMMA); !succ {
			break
		}
	}

	if succ, toks := p.accept(token.IN); !succ {
		return p.errorStmt(true, "Invalid token in for: %v", toks[len(toks)-1])
	}

	in, err := p.parseExpr()
	if err {
		return in.(ast.Statement), true
	}
	f.In = in

	if succ, toks := p.accept(token.EOL, token.INDENT); !succ {
		return p.errorStmt(true, "Invalid token in for: %v", toks[len(toks)-1])
	}

	for p.peek().Type != token.DEDENT && p.peek().Type != token.EOF {
		st, _ := p.parseFuncStmt()
		f.AddStmt(st)
	}

	if succ, toks := p.accept(token.DEDENT, token.EOL); !succ {
		st, _ := p.errorStmt(true, "Invalid token in for: %v", toks[len(toks)-1])
		f.AddStmt(st)
	}

	return f, false
}

func (p *parser) parseLoopStmt(label string) (ast.Statement, bool) {
	if succ, toks := p.accept(token.LOOP, token.EOL, token.INDENT); !succ {
		return p.errorStmt(true, "Invalid token in loop: %v", toks[len(toks)-1])
	}

	l := &ast.Loop{Label: label}
	for p.peek().Type != token.DEDENT && p.peek().Type != token.EOF {
		st, _ := p.parseFuncStmt()
		l.AddStmt(st)
	}

	if succ, toks := p.accept(token.DEDENT, token.EOL); !succ {
		st, _ := p.errorStmt(true, "Invalid token in loop: %v", toks[len(toks)-1])
		l.AddStmt(st)
	}

	return l, false
}

func (p *parser) parseIfInnerStmt() (ast.Statement, bool) {
	switch p.peek().Type {
	case token.IS:
		return p.parseIsStmt()
	default:
		return p.parseFuncStmt()
	}
}

func (p *parser) parseIsStmt() (ast.Statement, bool) {
	p.next() // eat is
	cond, _ := p.parseExprList()
	if succ, toks := p.accept(token.EOL, token.INDENT); !succ {
		return p.errorStmt(true, "Invalid token in is statement: %v", toks[len(toks)-1])
	}

	iss := &ast.Is{Condition: cond}
	for p.peek().Type != token.DEDENT && p.peek().Type != token.EOF {
		st, _ := p.parseFuncStmt()
		iss.AddStmt(st)
	}

	if succ, toks := p.accept(token.DEDENT, token.EOL); !succ {
		st, _ := p.errorStmt(true, "Invalid token in is statement: %v", toks[len(toks)-1])
		iss.AddStmt(st)
	}

	return iss, false
}

func (p *parser) parseIfStmt() (ast.Statement, bool) {
	p.next() // eat if
	var cond ast.Expression
	var err bool
	if p.peek().Type != token.EOL && p.peek().Type != token.WITH {
		cond, err = p.parseExpr()
		if err {
			return cond.(ast.Statement), true
		}
	}
	var with ast.Statement
	if succ, _ := p.accept(token.WITH); succ {
		switch p.peek().Type {
		case token.VAR:
			with, _ = p.parseVarStmt(true)
		default:
			with, _ = p.parseExprStmt(true)
		}
	}
	if succ, toks := p.accept(token.EOL, token.INDENT); !succ {
		return p.errorStmt(true, "Invalid token in if statement: %v", toks[len(toks)-1])
	}

	ifs := &ast.If{Condition: cond, With: with}
	for p.peek().Type != token.DEDENT && p.peek().Type != token.EOF {
		st, _ := p.parseIfInnerStmt()
		ifs.AddStmt(st)
	}

	if succ, toks := p.accept(token.DEDENT, token.EOL); !succ {
		st, _ := p.errorStmt(true, "Invalid token in if statement: %v", toks[len(toks)-1])
		ifs.AddStmt(st)
	}

	return ifs, false
}

func (p *parser) parseExprStmt(inWith bool) (ast.Statement, bool) {
	ex, _ := p.parseExprList()
	var assign ast.Statement
	if p.peek().Type.IsAssign() {
		assign = p.parseAssignStmt(ex)
	}
	if !inWith {
		if succ, toks := p.accept(token.EOL); !succ {
			return p.errorStmt(true, "Invalid token in expression statement: %v", toks[len(toks)-1])
		}
	}
	if assign != nil {
		return assign, false
	}
	return &ast.ExprStmt{Expr: ex}, false
}

func (p *parser) parseAssignStmt(lhs ast.Expression) ast.Statement {
	op := p.next().Type
	rhs, _ := p.parseExprList()
	return &ast.Assign{Op: op, Left: lhs, Right: rhs}
}

func (p *parser) parseReturnStmt() (ast.Statement, bool) {
	p.next() // eat ret
	r := &ast.Return{}
	if p.peek().Type != token.EOL {
		r.Vals, _ = p.parseExprList()
	}
	if succ, toks := p.accept(token.EOL); !succ {
		r.Vals, _ = p.errorExpr(true, "Invalid token in return statement: %v", toks[len(toks)-1])
	}
	return r, false
}

func (p *parser) parseDeferStmt() (ast.Statement, bool) {
	p.next() // eat defer
	ex, _ := p.parseExpr()
	d := &ast.Defer{Expr: ex}
	if succ, toks := p.accept(token.EOL); !succ {
		d.Expr, _ = p.errorExpr(true, "Invalid token in defer statement: %v", toks[len(toks)-1])
	}
	return d, false
}

func (p *parser) parseVarStmt(inWith bool) (ast.Statement, bool) {
	p.next() // eat var
	vs := &ast.VarSet{}

	vsl, err := p.parseVarLineStmt(inWith)
	if err {
		return vsl, true
	}
	vs.AddLine(vsl.(*ast.VarSetLine))

	if !inWith {
		if succ, _ := p.accept(token.INDENT); succ {
			for p.peek().Type != token.DEDENT && p.peek().Type != token.EOF {
				vsl, err = p.parseVarLineStmt(inWith)
				if err {
					return vsl, true
				}
				vs.AddLine(vsl.(*ast.VarSetLine))
			}

			if succ, toks := p.accept(token.DEDENT, token.EOL); !succ {
				return p.errorStmt(true, "Invalid token in var statement: %v", toks[len(toks)-1])
			}
		}
	}

	return vs, false
}

func (p *parser) parseVarLineStmt(inWith bool) (ast.Statement, bool) {
	v := &ast.VarSetLine{}
	for {
		if p.peek().Type != token.IDENTIFIER && p.peek().Type != token.BLANK {
			return p.errorStmt(true, "Invalid token in var statement: %v", p.peek())
		}
		name := p.next().Val

		var typ ast.Statement
		if p.peek().Type.IsType() {
			typ, _ = p.parseType()
		}

		v.AddVar(name, typ)
		if succ, _ := p.accept(token.COMMA); !succ {
			break
		}
	}

	if succ, _ := p.accept(token.ASSIGN); succ {
		v.Vals, _ = p.parseExprList()
	}

	if !inWith {
		if succ, toks := p.accept(token.EOL); !succ {
			return p.errorStmt(true, "Invalid token in var statement: %v", toks[len(toks)-1])
		}
	}
	return v, false
}

func (p *parser) parseExprList() (ast.Expression, bool) {
	el := &ast.ExprList{}
loop:
	for {
		e, err := p.parseExpr()
		el.AddExpr(e)
		if err {
			break loop
		}
		if succ, _ := p.accept(token.COMMA); !succ {
			break loop
		}
	}
	return el, false
}

func (p *parser) parseMLExprList(start, end token.Type) (ast.Expression, bool) {
	el := &ast.ExprList{}
	if succ, toks := p.accept(start); !succ {
		ex, _ := p.errorExpr(true, "Invalid token in expression list: %v", toks[len(toks)-1])
		el.AddExpr(ex)
		return el, true
	}
	switch p.peek().Type {
	case end: // do nothing
	default:
		if succ, _ := p.accept(token.EOL, token.INDENT); succ {
		loop:
			for {
				e, err := p.parseExpr()
				el.AddExpr(e)
				if err {
					break loop
				}
				if succ, _ := p.accept(token.EOL, token.DEDENT, token.EOL); succ {
					break loop
				}
				if succ, toks := p.accept(token.COMMA); !succ {
					ex, _ := p.errorExpr(true, "Invalid token in expression list: %v", toks[len(toks)-1])
					el.AddExpr(ex)
					break loop
				}
				p.accept(token.EOL) // eat EOL if it's there
			}
		} else {
			ex, err := p.parseExprList()
			if err {
				el.AddExpr(ex)
				return el, true
			}
			el = ex.(*ast.ExprList)
		}
	}
	if succ, toks := p.accept(end); !succ {
		ex, _ := p.errorExpr(true, "Invalid token in expression list: %v", toks[len(toks)-1])
		el.AddExpr(ex)
	}
	return el, false
}

func (p *parser) parseClassStmtIdent() (ast.Statement, bool) {
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
		case t == token.LEFT_PAREN:
			return p.parseFuncDef(dotted, name)
		case t.IsType():
			typ, _ = p.parseType()
		}

		ps.AddProp(!dotted, name, typ)
		if succ, _ = p.accept(token.COMMA); !succ {
			break
		}
	}

	if succ, _ := p.accept(token.ASSIGN); succ {
		ps.Vals, _ = p.parseExprList()
	}

	if succ, toks := p.accept(token.EOL); !succ {
		return p.errorStmt(true, "Invalid token in class statement: %v", toks[len(toks)-1])
	}

	return ps, false
}

func (p *parser) parseReturnValues() ([]ast.Statement, bool) {
	rvs := make([]ast.Statement, 0, 1)

	switch t := p.peek().Type; {
	case t.IsType():
		st, _ := p.parseType()
		rvs = append(rvs, st)
	case t == token.LEFT_PAREN:
		p.next() // eat (
		for p.peek().Type != token.RIGHT_PAREN {
			st, _ := p.parseType()
			rvs = append(rvs, st)
			switch p.peek().Type {
			case token.COMMA:
				p.next() // eat ,
			case token.RIGHT_PAREN:
			default:
				st, _ := p.errorStmt(true, "Invalid token in return types: %v", p.peek())
				return append(rvs, st), true
			}
		}
		p.next() // eat )
	}

	return rvs, false
}

func (p *parser) parseAnonFuncExpr() (ast.Expression, bool) {
	p.next() // eat fn
	st, err := p.parseFuncDef(true, "")
	return st.(ast.Expression), err
}

// parseFuncDef parses a function definition with the optional dot and name
// already consumed.
func (p *parser) parseFuncDef(dotted bool, name string) (ast.Statement, bool) {
	if succ, toks := p.accept(token.LEFT_PAREN); !succ {
		return p.errorStmt(true, "Invalid token in function definition: %v", toks[len(toks)-1])
	}
	f := &ast.FunctionDef{Static: !dotted, Name: name}
	for p.peek().Type != token.RIGHT_PAREN {
		succ, toks := p.accept(token.IDENTIFIER)
		if !succ {
			return p.errorStmt(true, "Invalid token in function definition: %v", p.peek())
		}
		name := toks[0].Val
		var typ ast.Statement
		if p.peek().Type.IsType() {
			typ, _ = p.parseType()
		}
		f.AddParam(name, typ)
		switch p.peek().Type {
		case token.COMMA:
			p.next() // eat ,
		case token.RIGHT_PAREN:
		default:
			return p.errorStmt(true, "Invalid token in function definition: %v", p.peek())
		}
	}
	if succ, toks := p.accept(token.RIGHT_PAREN); !succ {
		return p.errorStmt(true, "Invalid token in function definition: %v", toks[len(toks)-1])
	}

	// return value(s)
	rvs, _ := p.parseReturnValues()
	for _, rv := range rvs {
		f.AddReturn(rv)
	}

	if succ, toks := p.accept(token.EOL, token.INDENT); !succ {
		return p.errorStmt(true, "Invalid token in function definition: %v", toks[len(toks)-1])
	}

	for p.peek().Type != token.DEDENT && p.peek().Type != token.EOF {
		st, _ := p.parseFuncStmt()
		f.AddStmt(st)
	}

	if succ, toks := p.accept(token.DEDENT, token.EOL); !succ {
		st, _ := p.errorStmt(true, "Invalid token in function definition: %v", toks[len(toks)-1])
		f.AddStmt(st)
	}

	// If it's an anonymous function and we're not in the middle of a block
	// (followed by either a ',' or ')' ) then put the EOL back.
	if name == "" && !p.peek().Type.IsInBlock() {
		p.backup(1)
	}

	return f, false
}

func (p *parser) parsePrimaryExpr() (ast.Expression, bool) {
	var lhs ast.Expression
	switch p.peek().Type {
	case token.LEFT_PAREN:
		lhs, _ = p.parseParenExpr()
	case token.LEFT_CURLY:
		lhs, _ = p.parseCurlyExpr()
	case token.LEFT_BRACKET:
		lhs, _ = p.parseArrayCons()
	case token.IDENTIFIER:
		lhs, _ = p.parseIdentExpr()
	case token.IOTA:
		lhs, _ = p.parseIotaExpr()
	case token.BLANK:
		lhs, _ = p.parseBlankExpr()
	case token.STRING:
		lhs, _ = p.parseStringExpr()
	case token.NUMBER:
		lhs, _ = p.parseNumberExpr()
	case token.CHAR:
		lhs, _ = p.parseCharExpr()
	case token.TRUE, token.FALSE:
		lhs, _ = p.parseBoolExpr()
	case token.FUNCTION:
		lhs, _ = p.parseAnonFuncExpr()
	default:
		if p.peek().Type.IsUnaryOp() {
			lhs, _ = p.parseUnaryExpr()
		}
	}

	if lhs != nil {
		// Function calls, constructors and accessors.
	loop:
		for {
			switch p.peek().Type {
			case token.LEFT_BRACKET:
				lhs, _ = p.parseAccessorStmt(lhs)
			case token.LEFT_CURLY:
				lhs, _ = p.parseConstructor(lhs)
			case token.LEFT_PAREN:
				lhs, _ = p.parseFuncCallStmt(lhs)
			default:
				break loop
			}
		}
		return lhs, false
	}

	return p.errorExpr(true, "Token is not an expression: %v", p.peek())
}

func (p *parser) parseConstructor(lhs ast.Expression) (ast.Expression, bool) {
	p.next() // eat {
	con := &ast.Constructor{Type: lhs}
	switch p.peek().Type {
	case token.RIGHT_CURLY: // do nothing
	default:
		if succ, _ := p.accept(token.EOL, token.INDENT); succ {
		l1:
			for {
				kv, err := p.parseKeyVal()
				con.AddParam(kv)
				if err {
					break l1
				}
				if succ, _ := p.accept(token.EOL, token.DEDENT, token.EOL); succ {
					break l1
				}
				if succ, toks := p.accept(token.COMMA); !succ {
					return p.errorExpr(true, "Invalid token in constructor: %v", toks[len(toks)-1])
				}
				p.accept(token.EOL) // eat EOL if it's there
			}
		} else {
		l2:
			for {
				kv, err := p.parseKeyVal()
				con.AddParam(kv)
				if err {
					break l2
				}
				if succ, _ := p.accept(token.COMMA); !succ {
					break l2
				}
			}
		}
	}
	if succ, toks := p.accept(token.RIGHT_CURLY); !succ {
		return p.errorExpr(true, "Invalid token in constructor: %v", toks[len(toks)-1])
	}
	return con, false
}

func (p *parser) parseKeyVal() (ast.Statement, bool) {
	succ, toks := p.accept(token.IDENTIFIER, token.COLON)
	if !succ {
		return p.errorStmt(true, "Invalid token in key:value pair: %v", toks[len(toks)-1])
	}
	kv := &ast.KeyVal{Key: toks[0].Val}
	ex, err := p.parseExpr()
	kv.Val = ex
	return kv, err
}

func (p *parser) parseArrayCons() (ast.Expression, bool) {
	p.next() // eat [
	size, err := p.parseExpr()
	if err {
		return size, true
	}
	if succ, toks := p.accept(token.RIGHT_BRACKET); !succ {
		return p.errorExpr(true, "Invalid token in array constructor: %v", toks[len(toks)-1])
	}
	typ, err := p.parseType()
	if err {
		return typ.(ast.Expression), true
	}
	return &ast.ArrayCons{Type: typ, Size: size}, false
}

func (p *parser) parseCurlyExpr() (ast.Expression, bool) {
	ex, err := p.parseMLExprList(token.LEFT_CURLY, token.RIGHT_CURLY)
	return &ast.ArrayValueList{Vals: ex}, err
}

func (p *parser) parseAccessorStmt(lhs ast.Expression) (ast.Expression, bool) {
	p.next() // eat [

	var low, high ast.Expression
	var err bool
	isRange := false

	if p.peek().Type != token.COLON {
		low, err = p.parseExpr()
		if err {
			return low, true
		}
	}

	if succ, _ := p.accept(token.COLON); succ {
		isRange = true
		if p.peek().Type != token.RIGHT_BRACKET {
			high, err = p.parseExpr()
			if err {
				return high, true
			}
		}
	}

	if succ, toks := p.accept(token.RIGHT_BRACKET); !succ {
		return p.errorExpr(true, "Invalid token in accessor: %v", toks[len(toks)-1])
	}

	if isRange {
		return &ast.AccessorRange{Object: lhs, Low: low, High: high}, false
	}
	return &ast.Accessor{Object: lhs, Index: low}, false
}

func (p *parser) parseFuncCallStmt(lhs ast.Expression) (ast.Expression, bool) {
	fc := &ast.FunctionCall{Function: lhs}
	fc.Params, _ = p.parseMLExprList(token.LEFT_PAREN, token.RIGHT_PAREN)
	return fc, false
}

func (p *parser) parseUnaryExpr() (ast.Expression, bool) {
	op := p.next().Type
	ex, _ := p.parsePrimaryExpr()
	return &ast.Unary{Expr: ex, Op: op}, false
}

func (p *parser) parseParenExpr() (ast.Expression, bool) {
	p.next() // eat (
	expr, _ := p.parseExpr()
	if succ, toks := p.accept(token.RIGHT_PAREN); !succ {
		return p.errorExpr(true, "Invalid token in (): %v", toks[len(toks)-1])
	}
	return expr, false
}

func (p *parser) parseIdentExpr() (ast.Expression, bool) {
	ie := &ast.Identifier{}
	for {
		succ, toks := p.accept(token.IDENTIFIER)
		if !succ {
			return p.errorExpr(true, "Invalid token in identifier: %v", toks[len(toks)-1])
		}
		ip := &ast.IdentPart{Name: toks[0].Val}
		if succ, _ := p.accept(token.LEFT_CARET); succ {
			resetPos := p.pos - 1 // store the position in case the caret isn't a generic
			for p.peek().Type.IsType() {
				st, _ := p.parseType()
				ip.AddTypeParam(st)
				if succ, _ = p.accept(token.COMMA); !succ {
					break
				}
			}
			if succ, _ := p.accept(token.RIGHT_CARET); !succ {
				p.pos = resetPos     // no closing caret, so reset position
				ip.ResetTypeParams() // and reset type parameters
			}
		}
		ie.AddIdent(ip)
		if succ, _ := p.accept(token.DOT); !succ {
			break
		}
	}
	return ie, false
}

func (p *parser) parseBoolExpr() (ast.Expression, bool) {
	return &ast.Bool{Val: p.next().Type == token.TRUE}, false
}

func (p *parser) parseCharExpr() (ast.Expression, bool) {
	return &ast.Char{Val: p.next().Val}, false
}

func (p *parser) parseNumberExpr() (ast.Expression, bool) {
	return &ast.Number{Val: p.next().Val}, false
}

func (p *parser) parseIotaExpr() (ast.Expression, bool) {
	p.next() // eat iota
	return &ast.Iota{}, false
}

func (p *parser) parseBlankExpr() (ast.Expression, bool) {
	p.next() // eat _
	return &ast.Blank{}, false
}

func (p *parser) parseStringExpr() (ast.Expression, bool) {
	return &ast.String{Val: p.next().Val}, false
}

func (p *parser) parseExpr() (ast.Expression, bool) {
	lhs, err := p.parsePrimaryExpr()
	if err {
		return lhs, true
	}
	return p.parseBinopRHS(0, lhs)
}

func (p *parser) parseBinopRHS(exprPrec int, lhs ast.Expression) (ast.Expression, bool) {
	for {
		tokPrec := p.peekCombo().Precedence()

		// If this is a binary operator that binds as tightly as the
		// current one, consume it. Otherwise we're done.
		if tokPrec < exprPrec {
			return lhs, false
		}

		op := p.nextCombo()

		rhs, err := p.parsePrimaryExpr()
		if err {
			return rhs, true // An error, so rhs should hold the error message
		}

		// If binop binds less tightly with RHS than the operator after RHS,
		// let the pending op take RHS as its LHS.
		nextPrec := p.peekCombo().Precedence()
		if tokPrec < nextPrec {
			rhs, err = p.parseBinopRHS(tokPrec+1, rhs)
			if err {
				return rhs, true // An error, so rhs should hold the error message
			}
		}

		// Merge lhs/rhs
		lhs = &ast.Binary{Op: op.Type, Left: lhs, Right: rhs}
	}
}

func (p *parser) parseIotaStmt() (ast.Statement, bool) {
	if succ, toks := p.accept(token.IOTA, token.EOL); !succ {
		return p.errorStmt(true, "Invalid token in iota reset: %v", toks[len(toks)-1])
	}
	return &ast.Iota{}, false
}

func (p *parser) parseTypeIdent() (ast.Statement, bool) {
	t := &ast.TypeIdent{}
	t.AddIdent(p.next().Val)
	for {
		succ, toks := p.accept(token.DOT, token.IDENTIFIER)
		if !succ {
			break
		}
		t.AddIdent(toks[1].Val)
	}
	if succ, _ := p.accept(token.LEFT_CARET); succ {
		for p.peek().Type.IsType() {
			st, _ := p.parseType()
			t.AddTypeParam(st)
			if succ, _ = p.accept(token.COMMA); !succ {
				break
			}
		}
		if succ, toks := p.accept(token.RIGHT_CARET); !succ {
			return p.errorStmt(true, "Invalid token parsing type identifier: %v", toks[len(toks)-1])
		}
	}
	return t, false
}

func (p *parser) parseUse() (ast.Statement, bool) {
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
		if succ, _ = p.accept(token.DEDENT, token.EOL); succ {
			return u, false
		}
		return p.errorStmt(true, "Invalid token found when parsing Use: %v", errTok)
	}

	return u, false
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
