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
	fmt.Println(g.code)
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
