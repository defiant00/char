/*
TODO

Parsing:
	for
	loop
	loop labels
	break
	all operators
	proper operator order
	arrays
	constructors
	accessor
	unary op(s)
*/

package main

import (
	"fmt"
	"github.com/defiant00/char/compiler"
	"os"
)

func main() {
	fmt.Println("Char Compiler v0.1")
	var build, format, printTokens, printAST bool
	if len(os.Args) < 2 {
		fmt.Println("Usage: char <path> [parameters]")
		return
	}
	path := os.Args[1]
	for _, a := range os.Args[2:] {
		switch a {
		case "-build":
			build = true
		case "-format":
			format = true
		case "-printTokens":
			printTokens = true
		case "-printAST":
			printAST = true
		default:
			fmt.Printf("Unknown parameter %v\n", a)
		}
	}
	compiler.Build(path, build, format, printTokens, printAST)
	fmt.Println("\nDone")
}
