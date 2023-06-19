package reader

import (
	"burp/burper"
	"fmt"
	"os"
)

func Read(file string) (*burper.Tree, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println("ERR burp failed to close burp.toml: ", err)
		}
	}(f)
	return burper.New(f)
}
