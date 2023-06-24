package api

import (
	"burp/internal/burper"
	"burp/internal/services"
	"burp/pkg/console"
	"burp/pkg/fileutils"
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

func GetEnvironmentFile(burp *services.Burp) (*[]byte, bool) {
	if burp.Environment.ServerSide {
		fmt.Println(chalk.Red, "⚠", chalk.Reset, "Server-side translation is enabled, skipping local translations of environment file.")
		return nil, true
	}
	translation, err := burp.Environment.Translate("")
	if err != nil {
		fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We couldn't read the environment file!")
		fmt.Println(chalk.Red, err.Error())
		return nil, false
	}
	if burp.Environment.Override {
		fmt.Println(chalk.Red, "⚠", chalk.Reset, "Environment file override is enabled, overriding ", console.Highlight,
			burp.Environment.Baseline, chalk.Reset)
		if err = fileutils.Save(burp.Environment.Baseline, []byte(*translation)); err != nil {
			fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We couldn't override the environment file!")
			fmt.Println(chalk.Red, err.Error())
			return nil, false
		}
	}
	return utils.Ptr([]byte(*translation)), true
}
