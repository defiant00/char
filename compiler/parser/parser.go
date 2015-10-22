package parser

import (
	"fmt"
	"github.com/defiant00/char/compiler/lexer"
	"io/ioutil"
)

func Parse(file string) {
	fmt.Println("Parsing file", file)

	dat, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	input := string(dat)

	fmt.Println(input)

	fmt.Println("Data loaded...")
}
