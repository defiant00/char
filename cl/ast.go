package cl

import (
	"fmt"
)

type exprAST interface {
	Print(int)
	GenGo() string
}

func printSpaces(count int) {
	for i := 0; i < count; i++ {
		fmt.Print("  ")
	}
}

type errorAST struct {
	error string
}

func (e errorAST) Print(indent int) {
	printSpaces(indent)
	fmt.Println("Error:", e.error)
}

func (e errorAST) GenGo() string {
	return "\n// " + e.error
}

type programAST struct {
	items []exprAST
}

func (p *programAST) add(i exprAST) {
	p.items = append(p.items, i)
}

func (p programAST) Print(indent int) {
	printSpaces(indent)
	fmt.Println("program")
	for _, e := range p.items {
		e.Print(indent + 1)
	}
}

func (p programAST) GenGo() string {
	var s string
	for _, i := range p.items {
		s += i.GenGo() + "\n"
	}
	return s
}

type goBlockAST struct {
	code string
}

func (g goBlockAST) Print(indent int) {
	printSpaces(indent)
	fmt.Println("[go code]")
}

func (g goBlockAST) GenGo() string {
	return g.code
}

type packageAST struct {
	name string
}

func (p packageAST) Print(indent int) {
	printSpaces(indent)
	fmt.Printf("package %v\n", p.name)
}

func (p packageAST) GenGo() string {
	return fmt.Sprintf("package %v", p.name)
}

type numberAST struct {
	number string
}

func (n numberAST) Print(indent int) {
	printSpaces(indent)
	fmt.Printf("number %v\n", n.number)
}

func (n numberAST) GenGo() string {
	return n.number
}

type variableAST struct {
	name string
}

func (v variableAST) Print(indent int) {
	printSpaces(indent)
	fmt.Printf("variable %v\n", v.name)
}

func (v variableAST) GenGo() string {
	return v.name
}

type binaryExprAST struct {
	op          tType
	left, right exprAST
}

func (b binaryExprAST) Print(indent int) {
	b.left.Print(indent + 1)
	fmt.Println()
	printSpaces(indent)
	fmt.Println(b.op)
	b.right.Print(indent + 1)
}

func (b binaryExprAST) GenGo() string {
	return b.left.GenGo() + " HAHAHA OPERATOR GOES HERE " + b.right.GenGo()
}
