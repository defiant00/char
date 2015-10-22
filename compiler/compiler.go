package compiler

import (
	"fmt"
	"github.com/defiant00/char/compiler/parser"
	"os"
	"path/filepath"
)

func Build(path string) {
	fmt.Println("Building", path)

	d, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer d.Close()

	files, err := d.Readdir(-1)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.Mode().IsRegular() && filepath.Ext(file.Name()) == ".char" {
			parser.Parse(filepath.Join(path, file.Name()))
		}
	}
	/*
		file := parse(input)
		fmt.Println("\n")
		file.Print(0)
		fmt.Println("\n\nSaving .go file...")
		err = ioutil.WriteFile("../test/test.go", []byte(file.GenGo()), 0644)
		if err != nil {
			panic(err)
		}
	*/
}
