package reader

import (
	"burp/burper"
	"fmt"
	"os"
)

func Open(file string) (*os.File, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func Close(f *os.File) {
	err := f.Close()
	if err != nil {
		fmt.Println("ERR burp failed to close burp.toml: ", err)
	}
}

func Read(file string) (*burper.Tree, error) {
	f, err := Open(file)
	if err != nil {
		return nil, err
	}
	defer Close(f)
	return burper.New(f)
}
