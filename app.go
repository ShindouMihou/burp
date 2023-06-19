package main

import (
	"burp/burper/functions"
	"burp/reader"
	"fmt"
	"log"
)

func main() {
	functions.RegisterFunctions()
	tree, err := reader.Read("burp.toml")
	if err != nil {
		log.Fatalln(err)
		return
	}
	fmt.Println(tree.String())
}
