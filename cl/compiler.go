package cl

import (
	"fmt"
	"io/ioutil"
)

func Build(path string) {
	fmt.Println("Building", path)

	dat, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	input := string(dat)

	fmt.Println("Data loaded...")

	ast := parse(input)
	fmt.Println("\n")
	ast.Print(0)
}
