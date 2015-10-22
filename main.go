package main

import (
	"fmt"
	"github.com/defiant00/char/compiler"
)

func main() {
	fmt.Println("Char Compiler v0.01 Pre-Alpha")
	compiler.Build("c:/GoWorkspace/src/github.com/defiant00/char")
	fmt.Println("\nDone")
}
