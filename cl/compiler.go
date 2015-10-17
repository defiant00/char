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

	file := parse(input)
	fmt.Println("\n")
	file.Print(0)
	/*
		fmt.Println("\n\nSaving .go file")
		err = ioutil.WriteFile("test.go", []byte(file.GenGo()), 0644)
		if err != nil {
			panic(err)
		}
	*/
}
