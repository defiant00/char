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

func Parse(file string, printTokens bool) ast.General {
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
		p.fmtTokens = append(p.fmtTokens, t)
		if t.Type != token.SLCOMMENT {
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
		return p.errorStmt("\n\n%v\n", t)
	}
	return p.parseFile()
}

func (p *parser) errorStmt(format string, args ...interface{}) ast.Statement {
	return &ast.Error{Val: fmt.Sprintf(format, args...)}
}

func (p *parser) errorExpr(format string, args ...interface{}) ast.Expression {
	return &ast.Error{Val: fmt.Sprintf(format, args...)}
}

func (p *parser) peek() token.Token {
	return p.tokens[p.pos]
}

func (p *parser) next() token.Token {
	t := p.tokens[p.pos]
	p.pos++
	return t
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
		case token.USE:
			f.AddStmt(p.parseUse())
		default:
			return p.errorStmt("Unknown token %v", p.peek())
		}
	}
	return f
}

func (p *parser) parseUse() ast.Statement {
	p.next() // eat token.USE
	u := &ast.Use{}
	succ, toks := p.accept(token.EOL, token.INDENT)
	if succ {
		err, pack, alias, _ := p.parseUsePackage()
		for !err {
			u.AddPackage(pack, alias)
			err, pack, alias, _ = p.parseUsePackage()
		}
		succ, toks = p.accept(token.DEDENT)
		if succ {
			return u
		}
		return p.errorStmt("Unknown token found when parsing Use: %v", toks[0])
	}

	err, pack, alias, errTok := p.parseUsePackage()
	if err {
		return p.errorStmt("Unknown token found when parsing Use: %v", errTok)
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
