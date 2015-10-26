package ast

import (
	"fmt"
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
	Name       string
	typeParams []string
	withs      []*Identifier
	statements []Statement
}

func (c *Class) isStmt() {}

func (c *Class) Print(indent int) {
	printIndent(indent)
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

func (c *Class) AddWith(i *Identifier) {
	c.withs = append(c.withs, i)
}

func (c *Class) AddStmt(s Statement) {
	c.statements = append(c.statements, s)
}

type Identifier struct {
	idents []string
}

func (i Identifier) isStmt() {}

func (i *Identifier) Print(indent int) {
	fmt.Print(i)
}

func (i *Identifier) String() string {
	return strings.Join(i.idents, ".")
}

func (i *Identifier) AddIdent(ident string) {
	i.idents = append(i.idents, ident)
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

type Iota struct{}

func (i *Iota) isStmt() {}

func (i *Iota) Print(indent int) {
	printIndent(indent)
	fmt.Println("iota reset")
}

type Property struct {
	Static bool
	Name   string
	Type   Statement
}

func (p *Property) isStmt() {}

func (p *Property) Print(indent int) {
	printIndent(indent)
	if p.Static {
		fmt.Print("static ")
	}
	fmt.Print(p.Name)
	fmt.Println(p.Type)
}
