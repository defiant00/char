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
	for _, i := range p.items {
		i.Print(indent + 1)
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

type importAST struct {
	packages []string
}

func (i importAST) Print(indent int) {
	printSpaces(indent)
	fmt.Println("import")
	for _, p := range i.packages {
		printSpaces(indent + 1)
		fmt.Println(p)
	}
}

func (i importAST) GenGo() string {
	r := "import (\n"
	for _, p := range i.packages {
		r += "\"" + p + "\"\n"
	}
	r += ")"
	return r
}

type classAST struct {
	name  string
	items []exprAST
}

func (c classAST) Print(indent int) {
	printSpaces(indent)
	fmt.Printf("class %v\n", c.name)
	for _, i := range c.items {
		i.Print(indent + 1)
	}
}

func (c classAST) GenGo() string {
	var s string
	for _, i := range c.items {
		s += i.GenGo() + "\n"
	}
	return s
}

type funcDefAST struct {
	name   string
	static bool
}

func (f funcDefAST) Print(indent int) {
	printSpaces(indent)
	if f.static {
		fmt.Print("static ")
	}
	fmt.Println("function", f.name)
}

func (f funcDefAST) GenGo() string {
	return fmt.Sprintf("func %v()\n", f.name)
}

type varDefAST struct {
	name, typ string
	static    bool
}

func (v varDefAST) Print(indent int) {
	printSpaces(indent)
	if v.static {
		fmt.Print("static ")
	}
	fmt.Println("variable", v.name, v.typ)
}

func (v varDefAST) GenGo() string {
	return fmt.Sprintf("var %v %v", v.name, v.typ)
}
