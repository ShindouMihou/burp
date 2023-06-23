package api

import (
	"burp/internal/burper"
	"burp/internal/services"
	"burp/pkg/console"
	"burp/pkg/utils"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"github.com/ttacon/chalk"
	"path/filepath"
)

func GetBurper(directory string) (*services.Burp, *burper.Tree) {
	if !utils.HasSuffixStr(directory, ".toml") {
		directory = filepath.Join(directory, "burp.toml")
	}
	tree, err := burper.FromFile(directory)
	if err != nil {
		fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We couldn't analyze into ", console.Highlight, directory, " file!")
		fmt.Println(chalk.Red, err.Error())
		return nil, nil
	}
	var burp services.Burp
	if err = toml.Unmarshal(tree.Bytes(), &burp); err != nil {
		fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We couldn't analyze into ", console.Highlight, directory, " file!")
		fmt.Println(chalk.Red, err.Error())
		return nil, nil
	}
	return &burp, tree
}
