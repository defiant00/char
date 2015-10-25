package ast

import (
	"fmt"
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
