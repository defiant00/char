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
	t2 := p.peek()
	p.backup(1)
	if t.Type == token.RIGHT_CARET && t2.Type == token.RIGHT_CARET {
		return token.Token{Type: token.RSHIFT, Pos: t.Pos}
	}
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
			f.AddStmt(p.parseTopLevelIdent())
		case token.FUNCTION:
			f.AddStmt(p.parseTypeRedirect())
		case token.USE:
			f.AddStmt(p.parseUse())
		case token.INTERFACE:
			f.AddStmt(p.parseInterface())
		case token.MIXIN:
			f.AddStmt(p.parseMixin())
		default:
			f.AddStmt(p.errorStmt(true, "Invalid token %v", p.peek()))
		}
	}
	return f
}

func (p *parser) parseInterface() ast.Statement {
	succ, toks := p.accept(token.INTERFACE, token.IDENTIFIER)
	if !succ {
		return p.errorStmt(true, "Invalid token in interface: %v", toks[len(toks)-1])
	}

	name := toks[1].Val
	intf := &ast.InterfaceStmt{Name: name}

	if succ, toks = p.accept(token.EOL, token.INDENT); !succ {
		return p.errorStmt(true, "Invalid token in interface %v: %v", name, toks[len(toks)-1])
	}

	for p.peek().Type != token.DEDENT {
		// function_name(types)
		succ, toks = p.accept(token.IDENTIFIER, token.LEFT_PAREN)
		if !succ {
			return p.errorStmt(true, "Invalid token in interface %v: %v", name, toks[len(toks)-1])
		}
		fs := &ast.IntfFuncSig{Name: toks[0].Val}
		for p.peek().Type != token.RIGHT_PAREN {
			fs.AddParam(p.parseType())
			switch p.peek().Type {
			case token.COMMA:
				p.next() // eat ,
			case token.RIGHT_PAREN:
			default:
				return p.errorStmt(true, "Invalid token in interface %v function signature %v: %v", name, fs.Name, p.peek())
			}
		}
		p.next() // eat )

		// return value(s)
		rvs := p.parseReturnValues()
		for _, rv := range rvs {
			fs.AddReturn(rv)
		}

		if succ, _ = p.accept(token.EOL); !succ {
			return p.errorStmt(true, "Invalid token in interface %v function signature %v: %v", name, fs.Name, p.peek())
		}
		intf.AddFuncSig(fs)
	}

	if succ, toks = p.accept(token.DEDENT, token.EOL); !succ {
		return p.errorStmt(true, "Invalid token in interface %v: %v", name, toks[len(toks)-1])
	}

	return intf
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

func (p *parser) parseArrayType() ast.Statement {
	p.next() // eat []
	return &ast.ArrayType{Type: p.parseType()}
}

func (p *parser) parseFuncSigType() ast.Statement {
	f := &ast.FuncSigType{}

	// fn(types)
	if succ, toks := p.accept(token.FUNCTION, token.LEFT_PAREN); !succ {
		return p.errorStmt(true, "Invalid token in anonymous function: %v", toks[len(toks)-1])
	}
	for p.peek().Type != token.RIGHT_PAREN {
		f.AddParam(p.parseType())
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
	rvs := p.parseReturnValues()
	for _, rv := range rvs {
		f.AddReturn(rv)
	}

	return f
}

func (p *parser) parseClass(mixin bool) ast.Statement {
	succ, toks := p.accept(token.IDENTIFIER)
	if !succ {
		return p.errorStmt(true, "Invalid token in class declaration: %v", toks[len(toks)-1])
	}
	c := &ast.Class{Mixin: mixin, Name: toks[0].Val}

	if succ, _ := p.accept(token.LEFT_CARET); succ {
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
		if succ, _ = p.accept(token.RIGHT_CARET); !succ {
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

	if succ, toks := p.accept(token.DEDENT, token.EOL); !succ {
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

func (p *parser) parseBreakStmt() ast.Statement {
	p.next() // eat break
	b := &ast.Break{}
	if succ, toks := p.accept(token.IDENTIFIER); succ {
		b.Label = toks[0].Val
	}
	if succ, toks := p.accept(token.EOL); !succ {
		return p.errorStmt(true, "Invalid token in break: %v", toks[len(toks)-1])
	}
	return b
}

func (p *parser) parseForOrLoop(label string) ast.Statement {
	switch p.peek().Type {
	case token.FOR:
		return p.parseForStmt(label)
	case token.LOOP:
		return p.parseLoopStmt(label)
	default:
		return p.errorStmt(true, "Invalid token after label: %v", p.peek())
	}
}

func (p *parser) parseForStmt(label string) ast.Statement {
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

	f.In = p.parseExpr()
	switch f.In.(type) {
	case *ast.Error:
		return f.In.(ast.Statement)
	}

	if succ, toks := p.accept(token.EOL, token.INDENT); !succ {
		return p.errorStmt(true, "Invalid token in for: %v", toks[len(toks)-1])
	}

	for p.peek().Type != token.DEDENT && p.peek().Type != token.EOF {
		f.AddStmt(p.parseFuncStmt())
	}

	if succ, toks := p.accept(token.DEDENT, token.EOL); !succ {
		f.AddStmt(p.errorStmt(true, "Invalid token in for: %v", toks[len(toks)-1]))
	}

	return f
}

func (p *parser) parseLoopStmt(label string) ast.Statement {
	if succ, toks := p.accept(token.LOOP, token.EOL, token.INDENT); !succ {
		return p.errorStmt(true, "Invalid token in loop: %v", toks[len(toks)-1])
	}

	l := &ast.Loop{Label: label}
	for p.peek().Type != token.DEDENT && p.peek().Type != token.EOF {
		l.AddStmt(p.parseFuncStmt())
	}

	if succ, toks := p.accept(token.DEDENT, token.EOL); !succ {
		l.AddStmt(p.errorStmt(true, "Invalid token in loop: %v", toks[len(toks)-1]))
	}

	return l
}

func (p *parser) parseIfInnerStmt() ast.Statement {
	switch p.peek().Type {
	case token.IS:
		return p.parseIsStmt()
	default:
		return p.parseFuncStmt()
	}
}

func (p *parser) parseIsStmt() ast.Statement {
	p.next() // eat is
	cond := p.parseExprList()
	if succ, toks := p.accept(token.EOL, token.INDENT); !succ {
		return p.errorStmt(true, "Invalid token in is statement: %v", toks[len(toks)-1])
	}

	iss := &ast.Is{Condition: cond}
	for p.peek().Type != token.DEDENT && p.peek().Type != token.EOF {
		iss.AddStmt(p.parseFuncStmt())
	}

	if succ, toks := p.accept(token.DEDENT, token.EOL); !succ {
		iss.AddStmt(p.errorStmt(true, "Invalid token in is statement: %v", toks[len(toks)-1]))
	}

	return iss
}

func (p *parser) parseIfStmt() ast.Statement {
	p.next() // eat if
	var cond ast.Expression
	if p.peek().Type != token.EOL && p.peek().Type != token.WITH {
		cond = p.parseExpr()
		switch cond.(type) {
		case *ast.Error:
			return cond.(ast.Statement)
		}
	}
	var with ast.Statement
	if succ, _ := p.accept(token.WITH); succ {
		switch p.peek().Type {
		case token.VAR:
			with = p.parseVarStmt(true)
		default:
			with = p.parseExprStmt(true)
		}
	}
	if succ, toks := p.accept(token.EOL, token.INDENT); !succ {
		return p.errorStmt(true, "Invalid token in if statement: %v", toks[len(toks)-1])
	}

	ifs := &ast.If{Condition: cond, With: with}
	for p.peek().Type != token.DEDENT && p.peek().Type != token.EOF {
		ifs.AddStmt(p.parseIfInnerStmt())
	}

	if succ, toks := p.accept(token.DEDENT, token.EOL); !succ {
		ifs.AddStmt(p.errorStmt(true, "Invalid token in if statement: %v", toks[len(toks)-1]))
	}

	return ifs
}

func (p *parser) parseExprStmt(inWith bool) ast.Statement {
	ex := p.parseExprList()
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
		return assign
	}
	return &ast.ExprStmt{Expr: ex}
}

func (p *parser) parseAssignStmt(lhs ast.Expression) ast.Statement {
	op := p.next().Type
	rhs := p.parseExprList()
	return &ast.AssignStmt{Op: op, Left: lhs, Right: rhs}
}

func (p *parser) parseReturnStmt() ast.Statement {
	p.next() // eat ret
	r := &ast.ReturnStmt{}
	if p.peek().Type != token.EOL {
		r.Vals = p.parseExprList()
	}
	if succ, toks := p.accept(token.EOL); !succ {
		r.Vals = p.errorExpr(true, "Invalid token in return statement: %v", toks[len(toks)-1])
	}
	return r
}

func (p *parser) parseDeferStmt() ast.Statement {
	p.next() // eat defer
	d := &ast.DeferStmt{Expr: p.parseExpr()}
	if succ, toks := p.accept(token.EOL); !succ {
		d.Expr = p.errorExpr(true, "Invalid token in defer statement: %v", toks[len(toks)-1])
	}
	return d
}

func (p *parser) parseVarStmt(inWith bool) ast.Statement {
	p.next() // eat var
	vs := &ast.VarSet{}

	vsl := p.parseVarLineStmt(inWith)
	switch vsl.(type) {
	case *ast.Error:
		return vsl
	}
	vs.AddLine(vsl.(*ast.VarSetLine))

	if !inWith {
		if succ, _ := p.accept(token.INDENT); succ {
			for p.peek().Type != token.DEDENT && p.peek().Type != token.EOF {
				vsl = p.parseVarLineStmt(inWith)
				switch vsl.(type) {
				case *ast.Error:
					return vsl
				}
				vs.AddLine(vsl.(*ast.VarSetLine))
			}

			if succ, toks := p.accept(token.DEDENT, token.EOL); !succ {
				return p.errorStmt(true, "Invalid token in var statement: %v", toks[len(toks)-1])
			}
		}
	}

	return vs
}

func (p *parser) parseVarLineStmt(inWith bool) ast.Statement {
	v := &ast.VarSetLine{}
	for {
		if p.peek().Type != token.IDENTIFIER && p.peek().Type != token.BLANK {
			return p.errorStmt(true, "Invalid token in var statement: %v", p.peek())
		}
		name := p.next().Val

		var typ ast.Statement
		if p.peek().Type.IsType() {
			typ = p.parseType()
		}

		v.AddVar(name, typ)
		if succ, _ := p.accept(token.COMMA); !succ {
			break
		}
	}

	if succ, _ := p.accept(token.ASSIGN); succ {
		v.Vals = p.parseExprList()
	}

	if !inWith {
		if succ, toks := p.accept(token.EOL); !succ {
			return p.errorStmt(true, "Invalid token in var statement: %v", toks[len(toks)-1])
		}
	}
	return v
}

func (p *parser) parseExprList() *ast.ExprList {
	el := &ast.ExprList{}
loop:
	for {
		e := p.parseExpr()
		el.AddExpr(e)
		switch e.(type) {
		case *ast.Error:
			break loop
		}
		if succ, _ := p.accept(token.COMMA); !succ {
			break loop
		}
	}
	return el
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
		case t == token.LEFT_PAREN:
			return p.parseFuncDef(dotted, name)
		case t.IsType():
			typ = p.parseType()
		}

		ps.AddProp(!dotted, name, typ)
		if succ, _ = p.accept(token.COMMA); !succ {
			break
		}
	}

	if succ, _ := p.accept(token.ASSIGN); succ {
		ps.Vals = p.parseExprList()
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
	case t == token.LEFT_PAREN:
		p.next() // eat (
		for p.peek().Type != token.RIGHT_PAREN {
			rvs = append(rvs, p.parseType())
			switch p.peek().Type {
			case token.COMMA:
				p.next() // eat ,
			case token.RIGHT_PAREN:
			default:
				return append(rvs, p.errorStmt(true, "Invalid token in return types: %v", p.peek()))
			}
		}
		p.next() // eat )
	}

	return rvs
}

func (p *parser) parseAnonFuncExpr() ast.Expression {
	p.next() // eat fn
	return p.parseFuncDef(true, "").(ast.Expression)
}

// parseFuncDef parses a function definition with the optional dot and name
// already consumed.
func (p *parser) parseFuncDef(dotted bool, name string) ast.Statement {
	if succ, toks := p.accept(token.LEFT_PAREN); !succ {
		return p.errorStmt(true, "Invalid token in function definition: %v", toks[len(toks)-1])
	}
	f := &ast.FuncDef{Static: !dotted, Name: name}
	for p.peek().Type != token.RIGHT_PAREN {
		succ, toks := p.accept(token.IDENTIFIER)
		if !succ {
			return p.errorStmt(true, "Invalid token in function definition: %v", p.peek())
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
		case token.RIGHT_PAREN:
		default:
			return p.errorStmt(true, "Invalid token in function definition: %v", p.peek())
		}
	}
	if succ, toks := p.accept(token.RIGHT_PAREN); !succ {
		return p.errorStmt(true, "Invalid token in function definition: %v", toks[len(toks)-1])
	}

	// return value(s)
	rvs := p.parseReturnValues()
	for _, rv := range rvs {
		f.AddReturn(rv)
	}

	if succ, toks := p.accept(token.EOL, token.INDENT); !succ {
		return p.errorStmt(true, "Invalid token in function definition: %v", toks[len(toks)-1])
	}

	for p.peek().Type != token.DEDENT && p.peek().Type != token.EOF {
		f.AddStmt(p.parseFuncStmt())
	}

	if succ, toks := p.accept(token.DEDENT, token.EOL); !succ {
		f.AddStmt(p.errorStmt(true, "Invalid token in function definition: %v", toks[len(toks)-1]))
	}

	// If it's an anonymous function and we're not in the middle of a block
	// (followed by either a ',' or ')' ) then put the EOL back.
	if name == "" && !p.peek().Type.IsInBlock() {
		p.backup(1)
	}

	return f
}

func (p *parser) parsePrimaryExpr() ast.Expression {
	var lhs ast.Expression
	switch p.peek().Type {
	case token.LEFT_PAREN:
		lhs = p.parseParenExpr()
	case token.LEFT_CURLY:
		lhs = p.parseCurlyExpr()
	case token.LEFT_BRACKET:
		lhs = p.parseArrayCons()
	case token.IDENTIFIER:
		lhs = p.parseIdentExpr()
	case token.IOTA:
		lhs = p.parseIotaExpr()
	case token.BLANK:
		lhs = p.parseBlankExpr()
	case token.STRING:
		lhs = p.parseStringExpr()
	case token.NUMBER:
		lhs = p.parseNumberExpr()
	case token.CHAR:
		lhs = p.parseCharExpr()
	case token.TRUE, token.FALSE:
		lhs = p.parseBoolExpr()
	case token.FUNCTION:
		lhs = p.parseAnonFuncExpr()
	default:
		if p.peek().Type.IsUnaryOp() {
			lhs = p.parseUnaryExpr()
		}
	}

	if lhs != nil {
		// Function calls and accessors.
	loop:
		for {
			switch p.peek().Type {
			case token.LEFT_BRACKET:
				lhs = p.parseAccessorStmt(lhs)
			case token.LEFT_PAREN:
				lhs = p.parseFuncCallStmt(lhs)
			default:
				break loop
			}
		}
		return lhs
	}

	return p.errorExpr(true, "Token is not an expression: %v", p.peek())
}

func (p *parser) parseArrayCons() ast.Expression {
	p.next() // eat [
	size := p.parseExpr()
	switch size.(type) {
	case *ast.Error:
		return size
	}
	if succ, toks := p.accept(token.RIGHT_BRACKET); !succ {
		return p.errorExpr(true, "Invalid token in array constructor: %v", toks[len(toks)-1])
	}
	typ := p.parseType()
	switch typ.(type) {
	case *ast.Error:
		return typ.(ast.Expression)
	}
	return &ast.ArrayCons{Type: typ, Size: size}
}

func (p *parser) parseCurlyExpr() ast.Expression {
	p.next() // eat {
	avl := &ast.ArrayValueList{}
	switch p.peek().Type {
	case token.RIGHT_CURLY: // do nothing
	default:
		if succ, _ := p.accept(token.EOL, token.INDENT); succ {
			avl.Vals = &ast.ExprList{}
		loop:
			for {
				e := p.parseExpr()
				avl.Vals.AddExpr(e)
				switch e.(type) {
				case *ast.Error:
					break loop
				}
				if succ, _ := p.accept(token.EOL, token.DEDENT, token.EOL); succ {
					break loop
				}
				if succ, toks := p.accept(token.COMMA, token.EOL); !succ {
					avl.Vals.AddExpr(p.errorExpr(true, "Invalid token in array value list: %v", toks[len(toks)-1]))
					break loop
				}
			}
		} else {
			avl.Vals = p.parseExprList()
		}
	}
	if succ, toks := p.accept(token.RIGHT_CURLY); !succ {
		return p.errorExpr(true, "Invalid token in array value list: %v", toks[len(toks)-1])
	}
	return avl
}

func (p *parser) parseAccessorStmt(lhs ast.Expression) ast.Expression {
	p.next() // eat [

	var low, high ast.Expression
	isRange := false

	if p.peek().Type != token.COLON {
		low = p.parseExpr()
		switch low.(type) {
		case *ast.Error:
			return low
		}
	}

	if succ, _ := p.accept(token.COLON); succ {
		isRange = true
		if p.peek().Type != token.RIGHT_BRACKET {
			high = p.parseExpr()
			switch high.(type) {
			case *ast.Error:
				return high
			}
		}
	}

	if succ, toks := p.accept(token.RIGHT_BRACKET); !succ {
		return p.errorExpr(true, "Invalid token in accessor: %v", toks[len(toks)-1])
	}

	if isRange {
		return &ast.AccessorRangeExpr{Object: lhs, Low: low, High: high}
	}
	return &ast.AccessorExpr{Object: lhs, Index: low}
}

func (p *parser) parseFuncCallStmt(lhs ast.Expression) ast.Expression {
	p.next() // eat (
	fc := &ast.FuncCallExpr{Function: lhs}
	if p.peek().Type != token.RIGHT_PAREN {
		fc.Params = p.parseExprList()
	}
	if succ, toks := p.accept(token.RIGHT_PAREN); !succ {
		return p.errorExpr(true, "Invalid token in function call: %v", toks[len(toks)-1])
	}
	return fc
}

func (p *parser) parseUnaryExpr() ast.Expression {
	op := p.next().Type
	return &ast.UnaryExpr{Expr: p.parsePrimaryExpr(), Op: op}
}

func (p *parser) parseParenExpr() ast.Expression {
	p.next() // eat (
	expr := p.parseExpr()
	if succ, toks := p.accept(token.RIGHT_PAREN); !succ {
		return p.errorExpr(true, "Invalid token in (): %v", toks[len(toks)-1])
	}
	return expr
}

func (p *parser) parseIdentExpr() ast.Expression {
	ie := &ast.IdentExpr{}
	for {
		succ, toks := p.accept(token.IDENTIFIER)
		if !succ {
			return p.errorExpr(true, "Invalid token in identifier: %v", toks[len(toks)-1])
		}
		ip := &ast.IdentPart{Name: toks[0].Val}
		if succ, _ := p.accept(token.LEFT_CARET); succ {
			resetPos := p.pos - 1 // store the position in case the caret isn't a generic
			for p.peek().Type.IsType() {
				ip.AddTypeParam(p.parseType())
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
	return ie
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
	p.next() // eat iota
	return &ast.IotaExpr{}
}

func (p *parser) parseBlankExpr() ast.Expression {
	p.next() // eat _
	return &ast.BlankExpr{}
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
		tokPrec := p.peekCombo().Precedence()

		// If this is a binary operator that binds as tightly as the
		// current one, consume it. Otherwise we're done.
		if tokPrec < exprPrec {
			return lhs
		}

		op := p.nextCombo()

		rhs := p.parsePrimaryExpr()
		switch rhs.(type) {
		case *ast.Error:
			return rhs // An error, so rhs should hold the error message
		}

		// If binop binds less tightly with RHS than the operator after RHS,
		// let the pending op take RHS as its LHS.
		nextPrec := p.peekCombo().Precedence()
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
	if succ, _ := p.accept(token.LEFT_CARET); succ {
		for p.peek().Type.IsType() {
			t.AddTypeParam(p.parseType())
			if succ, _ = p.accept(token.COMMA); !succ {
				break
			}
		}
		if succ, toks := p.accept(token.RIGHT_CARET); !succ {
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
		if succ, _ = p.accept(token.DEDENT, token.EOL); succ {
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
