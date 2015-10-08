package cl

import (
	"fmt"
)

type exprAST interface {
	Print(int)
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

type programAST struct {
	items []exprAST
}

func (p programAST) Print(indent int) {
	printSpaces(indent)
	fmt.Println("Program")
	for _, e := range p.items {
		e.Print(indent + 1)
	}
}

type packageAST struct {
	name string
}

func (p packageAST) Print(indent int) {
	printSpaces(indent)
	fmt.Printf("package %v\n", p.name)
}

type importAST struct {
	names []string
}

func (i importAST) Print(indent int) {
	printSpaces(indent)
	fmt.Println("import")
	for _, n := range i.names {
		printSpaces(indent + 1)
		fmt.Println(n)
	}
}
