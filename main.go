package main

import (
	"fmt"
	"github.com/defiant00/char/cl"
)

func main() {
	fmt.Println("Char Compiler v0.05 Pre-Alpha")
	cl.Build("test.char")
	fmt.Println("\nDone")
}
