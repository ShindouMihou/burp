package main

import (
	"burp/burper/functions"
	"burp/reader"
	"burp/services"
	"github.com/BurntSushi/toml"
	"log"
)

func main() {
	functions.RegisterFunctions()
	tree, err := reader.Read("burp.toml")
	if err != nil {
		log.Fatalln(err)
		return
	}
	var burp services.Burp
	_, err = toml.Decode(tree.String(), &burp)
	if err != nil {
		log.Fatalln(err)
		return
	}
	err = burp.Environment.Save()
	if err != nil {
		log.Fatalln(err)
		return
	}
}
