package cl

import (
	"fmt"
	"strings"
)

var currentClass string = ""

type genAST interface {
	Print(int)
	GenGo() string
}

type exprAST interface {
	genAST
}

func printIndent(count int) {
	for i := 0; i < count; i++ {
		fmt.Print("|   ")
	}
}

// The AST for an error.
type errorAST struct {
	error string
}

func (e errorAST) Print(indent int) {
	printIndent(indent)
	fmt.Println(e.error)
}

func (e *errorAST) GenGo() string {
	return fmt.Sprintf("// Error: %v", e.error)
}

// The AST for a source file.
type fileAST struct {
	statements []genAST
}

func (f fileAST) Print(indent int) {
	printIndent(indent)
	fmt.Println("file")
	for _, s := range f.statements {
		s.Print(indent + 1)
	}
}

func (f *fileAST) GenGo() string {
	ret := ""
	for _, s := range f.statements {
		ret += s.GenGo() + "\n"
	}
	return ret
}

func (f *fileAST) addStmt(s genAST) {
	f.statements = append(f.statements, s)
}

// The AST for a 'use' statement.
type useAST struct {
	packages []string
}

func (u useAST) Print(indent int) {
	printIndent(indent)
	fmt.Println("use")
	for _, p := range u.packages {
		printIndent(indent + 1)
		fmt.Println(p)
	}
}

func (u *useAST) GenGo() string {
	ret := "import (\n"
	for _, p := range u.packages {
		ret += "\t\"" + p + "\"\n"
	}
	ret += ")"
	return ret
}

func (u *useAST) addPkg(pack string) {
	u.packages = append(u.packages, pack)
}

// The AST for a Go block.
type goBlockAST struct {
	block string
}

func (g goBlockAST) Print(indent int) {
	printIndent(indent)
	fmt.Println("[Go Block]")
}

func (g *goBlockAST) GenGo() string {
	return g.block
}

// The AST for a class.
type classAST struct {
	name       string
	statements []genAST
}

func (c classAST) Print(indent int) {
	printIndent(indent)
	fmt.Println("class", c.name)
	for _, s := range c.statements {
		s.Print(indent + 1)
	}
}

func (c *classAST) GenGo() string {
	currentClass = c.name
	ret := ""
	for _, s := range c.statements {
		ret += s.GenGo() + "\n"
	}
	return ret
}

func (c *classAST) addStmt(stmt genAST) {
	c.statements = append(c.statements, stmt)
}

// The AST for a function definition.
type funcAST struct {
	name        string
	static      bool
	params      []paramAST
	returns     exprAST
	expressions []exprAST
}

func (f funcAST) Print(indent int) {
	printIndent(indent)
	if f.static {
		fmt.Print("static ")
	}
	fmt.Println("function", f.name)
	if len(f.params) > 0 {
		printIndent(indent + 1)
		fmt.Println("parameters")
		for _, p := range f.params {
			p.Print(indent + 2)
		}
	}
	if f.returns != nil {
		printIndent(indent + 1)
		fmt.Println("returns")
		f.returns.Print(indent + 2)
	}
	for _, e := range f.expressions {
		e.Print(indent + 1)
	}
}

func (f *funcAST) GenGo() string {
	n := currentClass + "_" + f.name
	if currentClass == "main" && f.name == "main" {
		n = "main"
	}
	ret := "func "
	if !f.static {
		ret += "(this *" + currentClass + ") "
	}
	ret += n + "("
	for i, p := range f.params {
		if i > 0 {
			ret += ","
		}
		ret += p.GenGo()
	}
	ret += ")"
	if f.returns != nil {
		ret += " (" + f.returns.GenGo() + ")"
	}
	ret += " {\n"
	for _, e := range f.expressions {
		ret += e.GenGo() + "\n"
	}
	ret += "}"
	return ret
}

func (f *funcAST) addParam(name string, typ exprAST) {
	f.params = append(f.params, paramAST{name: name, typ: typ})
}

func (f *funcAST) addExpr(expr exprAST) {
	f.expressions = append(f.expressions, expr)
}

// A function parameter.
type paramAST struct {
	name string
	typ  exprAST
}

func (p paramAST) Print(indent int) {
	printIndent(indent)
	fmt.Println(p.name)
	if p.typ != nil {
		p.typ.Print(indent + 1)
	}
}

func (p *paramAST) GenGo() string {
	ret := p.name
	if p.typ != nil {
		ret += " " + p.typ.GenGo()
	}
	return ret
}

// The AST for a 'var' declaration expression.
type varDeclareAST struct {
	initial exprAST
}

func (v varDeclareAST) Print(indent int) {
	printIndent(indent)
	fmt.Println("variable")
	v.initial.Print(indent + 1)
}

func (v *varDeclareAST) GenGo() string {
	return strings.Replace(v.initial.GenGo(), "=", ":=", 1)
}

// The AST for an identifier expression.
type identExprAST struct {
	name string
}

func (i identExprAST) Print(indent int) {
	printIndent(indent)
	fmt.Println("identifier", i.name)
}

func (i *identExprAST) GenGo() string {
	return i.name
}

// The AST for a string expression.
type stringExprAST struct {
	val string
}

func (s stringExprAST) Print(indent int) {
	printIndent(indent)
	fmt.Println("string", s.val)
}

func (s *stringExprAST) GenGo() string {
	return "\"" + s.val + "\""
}

// The AST for a boolean expression.
type boolExprAST struct {
	val bool
}

func (b boolExprAST) Print(indent int) {
	printIndent(indent)
	fmt.Println("bool", b.val)
}

func (b *boolExprAST) GenGo() string {
	if b.val {
		return "true"
	}
	return "false"
}

// The AST for a numeric expression.
type numberExprAST struct {
	val string
}

func (n numberExprAST) Print(indent int) {
	printIndent(indent)
	fmt.Println("number", n.val)
}

func (n *numberExprAST) GenGo() string {
	return n.val
}

// The AST for a binary expression.
type binaryExprAST struct {
	left, right exprAST
	op          tType
}

func (b binaryExprAST) Print(indent int) {
	printIndent(indent)
	fmt.Println(b.op)
	b.left.Print(indent + 1)
	b.right.Print(indent + 1)
}

func (b *binaryExprAST) GenGo() string {
	return b.left.GenGo() + tokGoOp[b.op] + b.right.GenGo()
}

// The AST for a class constant.
type constAST struct {
	name string
	val  exprAST
}

func (c constAST) Print(indent int) {
	printIndent(indent)
	fmt.Println("const", c.name)
	if c.val != nil {
		c.val.Print(indent + 1)
	}
}

func (c *constAST) GenGo() string {
	ret := "const " + c.name
	if c.val != nil {
		ret += " = " + c.val.GenGo()
	}
	return ret
}

// The AST for a class variable.
type propertyAST struct {
	name string
	typ  exprAST
}

func (p propertyAST) Print(indent int) {
	printIndent(indent)
	fmt.Println("property", p.name)
	p.typ.Print(indent + 1)
}

func (p *propertyAST) GenGo() string {
	return "class_prop_" + p.name + " " + p.typ.GenGo()
}

// The AST for a return statement.
type returnAST struct {
	val exprAST
}

func (r returnAST) Print(indent int) {
	printIndent(indent)
	fmt.Println("return")
	if r.val != nil {
		r.val.Print(indent + 1)
	}
}

func (r *returnAST) GenGo() string {
	ret := "return"
	if r.val != nil {
		ret += " " + r.val.GenGo()
	}
	return ret
}

// The AST for a function call.
type funcCallExprAST struct {
	name string
	args exprAST
}

func (f funcCallExprAST) Print(indent int) {
	printIndent(indent)
	fmt.Println("call func", f.name)
	if f.args != nil {
		f.args.Print(indent + 1)
	}
}

func (f *funcCallExprAST) GenGo() string {
	ret := f.name + "("
	if f.args != nil {
		ret += f.args.GenGo()
	}
	ret += ")"
	return ret
}
