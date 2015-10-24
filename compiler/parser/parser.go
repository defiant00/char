package parser

import (
	"fmt"
	"github.com/defiant00/char/compiler/lexer"
	"github.com/defiant00/char/compiler/token"
	"io/ioutil"
)

func Parse(file string, printTokens bool) {
	fmt.Println("Parsing file", file)

	dat, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	input := string(dat)
	fmt.Println("Data loaded...")

	l := lexer.Lex(input)
	var t token.Token

	// Read all tokens into a slice.
	for {
		t = l.NextToken()
		if printTokens {
			fmt.Print(" ", t)
		}
		if t.Type == token.ERROR || t.Type == token.EOF {
			break
		}
	}
	if t.Type == token.ERROR {
		fmt.Printf("\n\n%v\n", t)
	}
}
