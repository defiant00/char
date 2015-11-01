package ast

import (
	"fmt"
	"github.com/defiant00/char/compiler/token"
	"strings"
)

type General interface {
	Print(int)
}

type Statement interface {
	General
	isStmt()
}

type Expression interface {
	General
	isExpr()
}

func printIndent(indent int) {
	for i := 0; i < indent; i++ {
		fmt.Print("|   ")
	}
}

type Error struct {
	Val string
}

func (e *Error) isStmt() {}
func (e *Error) isExpr() {}

func (e *Error) Print(indent int) {
	printIndent(indent)
	fmt.Printf("ERROR: %v\n", e.Val)
}

type ExprStmt struct {
	Expr Expression
}

func (e *ExprStmt) isStmt() {}

func (e *ExprStmt) Print(indent int) {
	e.Expr.Print(indent)
}

type File struct {
	Name       string
	statements []Statement
}

func (f *File) Print(indent int) {
	printIndent(indent)
	fmt.Println(f.Name)
	for _, s := range f.statements {
		s.Print(indent + 1)
	}
}

func (f *File) AddStmt(stmt Statement) {
	f.statements = append(f.statements, stmt)
}

type Use struct {
	packages []usePackage
}

func (u *Use) isStmt() {}

func (u *Use) Print(indent int) {
	printIndent(indent)
	fmt.Println("use")
	for _, p := range u.packages {
		printIndent(indent + 1)
		fmt.Println(p)
	}
}

func (u *Use) AddPackage(pack, alias string) {
	u.packages = append(u.packages, usePackage{pack, alias})
}

type usePackage struct {
	pack, alias string
}

func (u usePackage) String() string {
	if u.alias != "" {
		return fmt.Sprintf("%v as %v", u.pack, u.alias)
	}
	return u.pack
}

type Class struct {
	Mixin      bool
	Name       string
	typeParams []string
	withs      []Statement
	statements []Statement
}

func (c *Class) isStmt() {}

func (c *Class) Print(indent int) {
	printIndent(indent)
	if c.Mixin {
		fmt.Print("mixin ")
	}
	fmt.Print("class ", c.Name)
	if len(c.typeParams) > 0 {
		fmt.Print("<", strings.Join(c.typeParams, ", "), ">")
	}
	if len(c.withs) > 0 {
		fmt.Print(" with ")
		for i, w := range c.withs {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(w)
		}
	}
	fmt.Println()
	for _, s := range c.statements {
		s.Print(indent + 1)
	}
}

func (c *Class) AddTypeParam(t string) {
	c.typeParams = append(c.typeParams, t)
}

func (c *Class) AddWith(s Statement) {
	c.withs = append(c.withs, s)
}

func (c *Class) AddStmt(s Statement) {
	c.statements = append(c.statements, s)
}

type TypeIdent struct {
	idents     []string
	typeParams []Statement
}

func (t *TypeIdent) isStmt() {}

func (t *TypeIdent) Print(indent int) {
	fmt.Print("TypeIdentifier ", t)
}

func (t *TypeIdent) String() string {
	ret := strings.Join(t.idents, ".")
	if len(t.typeParams) > 0 {
		ret += "<"
	}
	for i, tp := range t.typeParams {
		if i > 0 {
			ret += ", "
		}
		ret += fmt.Sprint(tp)
	}
	if len(t.typeParams) > 0 {
		ret += ">"
	}
	return ret
}

func (t *TypeIdent) AddIdent(ident string) {
	t.idents = append(t.idents, ident)
}

func (t *TypeIdent) AddTypeParam(s Statement) {
	t.typeParams = append(t.typeParams, s)
}

type TypeRedirect struct {
	Type Statement
	Name string
}

func (t *TypeRedirect) isStmt() {}

func (t *TypeRedirect) Print(indent int) {
	printIndent(indent)
	fmt.Printf("%v as %v\n", t.Type, t.Name)
}

type AnonFuncType struct {
	params  []Statement
	returns []Statement
}

func (a *AnonFuncType) isExpr() {}
func (a *AnonFuncType) isStmt() {}

func (a *AnonFuncType) Print(indent int) {
	printIndent(indent)
	fmt.Println(a)
}

func (a *AnonFuncType) String() string {
	ret := "func("
	for i, p := range a.params {
		if i > 0 {
			ret += ", "
		}
		ret += fmt.Sprint(p)
	}
	ret += ")"
	if len(a.returns) > 0 {
		ret += " "
		if len(a.returns) > 1 {
			ret += "("
		}
		for i, r := range a.returns {
			if i > 0 {
				ret += ", "
			}
			ret += fmt.Sprint(r)
		}
		if len(a.returns) > 1 {
			ret += ")"
		}
	}
	return ret
}

func (a *AnonFuncType) AddParam(p Statement) {
	a.params = append(a.params, p)
}

func (a *AnonFuncType) AddReturn(r Statement) {
	a.returns = append(a.returns, r)
}

type IotaStmt struct{}

func (i *IotaStmt) isStmt() {}

func (i *IotaStmt) Print(indent int) {
	printIndent(indent)
	fmt.Println("iota reset")
}

type PropertySet struct {
	props []property
	Vals  Expression
}

func (p *PropertySet) isStmt() {}

func (p *PropertySet) Print(indent int) {
	printIndent(indent)
	fmt.Print("prop set: ")
	for i, pr := range p.props {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(pr)
	}
	fmt.Println()
	if p.Vals != nil {
		p.Vals.Print(indent + 1)
	}
}

func (p *PropertySet) AddProp(static bool, name string, typ Statement) {
	p.props = append(p.props, property{static: static, name: name, typ: typ})
}

type property struct {
	static bool
	name   string
	typ    Statement
}

func (p property) String() string {
	var ret string
	if p.static {
		ret = "static "
	}
	ret += p.name
	if p.typ != nil {
		ret += fmt.Sprintf(" %v", p.typ)
	}
	return ret
}

type BinaryExpr struct {
	Left, Right Expression
	Op          token.Type
}

func (b *BinaryExpr) isExpr() {}

func (b *BinaryExpr) Print(indent int) {
	printIndent(indent)
	fmt.Println(b.Op)
	b.Left.Print(indent + 1)
	b.Right.Print(indent + 1)
}

type StringExpr struct {
	Val string
}

func (s *StringExpr) isExpr() {}

func (s *StringExpr) Print(indent int) {
	printIndent(indent)
	fmt.Printf("string '%v'\n", s.Val)
}

type NumberExpr struct {
	Val string
}

func (n *NumberExpr) isExpr() {}

func (n *NumberExpr) Print(indent int) {
	printIndent(indent)
	fmt.Println("number", n.Val)
}

type CharExpr struct {
	Val string
}

func (c *CharExpr) isExpr() {}

func (c *CharExpr) Print(indent int) {
	printIndent(indent)
	fmt.Printf("char '%v'\n", c.Val)
}

type BoolExpr struct {
	Val bool
}

func (b *BoolExpr) isExpr() {}

func (b *BoolExpr) Print(indent int) {
	printIndent(indent)
	fmt.Println("bool", b.Val)
}

type IotaExpr struct{}

func (i *IotaExpr) isExpr() {}

func (i *IotaExpr) Print(indent int) {
	printIndent(indent)
	fmt.Println("iota")
}

type IdentExpr struct {
	Idents []*IdentPart
}

func (i *IdentExpr) isExpr() {}

func (ie *IdentExpr) Print(indent int) {
	printIndent(indent)
	for i, id := range ie.Idents {
		if i > 0 {
			fmt.Print(".")
		}
		fmt.Print(id)
	}
	fmt.Println()
}

type FuncCallExpr struct {
	Idents []*IdentPart
	Params Expression
}

func (f *FuncCallExpr) isExpr() {}

func (f *FuncCallExpr) Print(indent int) {
	printIndent(indent)
	fmt.Print("func ")
	for i, id := range f.Idents {
		if i > 0 {
			fmt.Print(".")
		}
		fmt.Print(id)
	}
	fmt.Println()
	if f.Params != nil {
		f.Params.Print(indent + 1)
	}
}

type IdentPart struct {
	Name       string
	typeParams []Statement
}

func (ip *IdentPart) String() string {
	ret := ip.Name
	if len(ip.typeParams) > 0 {
		ret += "<"
	}
	for i, tp := range ip.typeParams {
		if i > 0 {
			ret += ", "
		}
		ret += fmt.Sprint(tp)
	}
	if len(ip.typeParams) > 0 {
		ret += ">"
	}
	return ret
}

func (i *IdentPart) AddTypeParam(s Statement) {
	i.typeParams = append(i.typeParams, s)
}

func (i *IdentPart) ResetTypeParams() {
	i.typeParams = make([]Statement, 0)
}

type FuncDefStmt struct {
	Static     bool
	Name       string
	params     []param
	returns    []Statement
	statements []Statement
}

func (f *FuncDefStmt) isStmt() {}

func (f *FuncDefStmt) Print(indent int) {
	printIndent(indent)
	if f.Static {
		fmt.Print("static ")
	}
	fmt.Print(f.Name)
	fmt.Print("(")
	for i, p := range f.params {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(p)
	}
	fmt.Print(")")

	if len(f.returns) > 0 {
		fmt.Print(" ")
		if len(f.returns) > 1 {
			fmt.Print("(")
		}
		for i, r := range f.returns {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(r)
		}
		if len(f.returns) > 1 {
			fmt.Print(")")
		}
	}

	fmt.Println()
	for _, s := range f.statements {
		s.Print(indent + 1)
	}
}

func (f *FuncDefStmt) AddStmt(s Statement) {
	f.statements = append(f.statements, s)
}

func (f *FuncDefStmt) AddParam(name string, typ Statement) {
	f.params = append(f.params, param{name: name, typ: typ})
}

func (f *FuncDefStmt) AddReturn(r Statement) {
	f.returns = append(f.returns, r)
}

type param struct {
	name string
	typ  Statement
}

func (p param) String() string {
	ret := p.name
	if p.typ != nil {
		ret += fmt.Sprint(" ", p.typ)
	}
	return ret
}

type VarSet struct {
	lines []*VarSetLine
}

func (v *VarSet) isStmt() {}

func (v *VarSet) Print(indent int) {
	printIndent(indent)
	fmt.Println("var set")
	for _, l := range v.lines {
		l.Print(indent + 1)
	}
}

func (v *VarSet) AddLine(vsl *VarSetLine) {
	v.lines = append(v.lines, vsl)
}

type VarSetLine struct {
	vars []variable
	Vals Expression
}

func (v *VarSetLine) isStmt() {}

func (v *VarSetLine) Print(indent int) {
	printIndent(indent)
	for i, vr := range v.vars {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(vr)
	}
	fmt.Println()
	if v.Vals != nil {
		v.Vals.Print(indent + 1)
	}
}

func (v *VarSetLine) AddVar(name string, typ Statement) {
	v.vars = append(v.vars, variable{name: name, typ: typ})
}

type variable struct {
	name string
	typ  Statement
}

func (v variable) String() string {
	ret := v.name
	if v.typ != nil {
		ret += fmt.Sprint(" ", v.typ)
	}
	return ret
}
