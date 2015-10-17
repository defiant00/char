package cl

import (
	"fmt"
)

type genAST interface {
	Print(int)
}

type exprAST interface {
	genAST
}

func printSpaces(count int) {
	for i := 0; i < count; i++ {
		fmt.Print("  ")
	}
}

// The AST for an error.
type errorAST struct {
	error string
}

func (e errorAST) Print(indent int) {
	printSpaces(indent)
	fmt.Println(e.error)
}

// The AST for a source file.
type fileAST struct {
	statements []genAST
}

func (f fileAST) Print(indent int) {
	printSpaces(indent)
	fmt.Println("file")
	for _, s := range f.statements {
		s.Print(indent + 1)
	}
}

func (f *fileAST) addStmt(s genAST) {
	f.statements = append(f.statements, s)
}

// The AST for a 'use' statement.
type useAST struct {
	packages []string
}

func (u useAST) Print(indent int) {
	printSpaces(indent)
	fmt.Println("use")
	for _, p := range u.packages {
		printSpaces(indent + 1)
		fmt.Println(p)
	}
}

func (u *useAST) addPkg(pack string) {
	u.packages = append(u.packages, pack)
}

// The AST for a Go block.
type goBlockAST struct {
	block string
}

func (g goBlockAST) Print(indent int) {
	printSpaces(indent)
	fmt.Println("[Go Block]")
}

// The AST for a class.
type classAST struct {
	name       string
	statements []genAST
}

func (c classAST) Print(indent int) {
	printSpaces(indent)
	fmt.Println("class", c.name)
	for _, s := range c.statements {
		s.Print(indent + 1)
	}
}

func (c *classAST) addStmt(stmt genAST) {
	c.statements = append(c.statements, stmt)
}

// The AST for a function definition.
type funcAST struct {
	name        string
	static      bool
	expressions []exprAST
}

func (f funcAST) Print(indent int) {
	printSpaces(indent)
	if f.static {
		fmt.Print("static ")
	}
	fmt.Println("function", f.name)
	for _, e := range f.expressions {
		e.Print(indent + 1)
	}
}

func (f *funcAST) addExpr(expr exprAST) {
	f.expressions = append(f.expressions, expr)
}

// The AST for a 'var' declaration expression.
type varDeclareAST struct {
	initial exprAST
}

func (v varDeclareAST) Print(indent int) {
	printSpaces(indent)
	fmt.Println("variable")
	v.initial.Print(indent + 1)
}

// The AST for an identifier expression.
type identExprAST struct {
	name string
}

func (i identExprAST) Print(indent int) {
	printSpaces(indent)
	fmt.Println("identifier", i.name)
}

// The AST for a string expression.
type stringExprAST struct {
	val string
}

func (s stringExprAST) Print(indent int) {
	printSpaces(indent)
	fmt.Println("string", s.val)
}

// The AST for a numeric expression.
type numberExprAST struct {
	val string
}

func (n numberExprAST) Print(indent int) {
	printSpaces(indent)
	fmt.Println("number", n.val)
}

// The AST for a binary expression.
type binaryExprAST struct {
	left, right exprAST
	op          tType
}

func (b binaryExprAST) Print(indent int) {
	printSpaces(indent)
	fmt.Println(b.op)
	b.left.Print(indent + 1)
	b.right.Print(indent + 1)
}

// The AST for a constant block.
type constAST struct {
	defs []exprAST
}

func (c constAST) Print(indent int) {
	printSpaces(indent)
	fmt.Println("constant")
	for _, d := range c.defs {
		d.Print(indent + 1)
	}
}

func (c *constAST) addDef(d exprAST) {
	c.defs = append(c.defs, d)
}

// The AST for a class variable.
type classVarAST struct {
	name string
	typ  exprAST
}

func (c classVarAST) Print(indent int) {
	printSpaces(indent)
	fmt.Println("class var", c.name)
	c.typ.Print(indent + 1)
}
