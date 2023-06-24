package api

import (
	"burp/internal/burp"
	"burp/internal/burper"
	"burp/pkg/console"
	"burp/pkg/fileutils"
	"burp/pkg/utils"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"github.com/ttacon/chalk"
	"path/filepath"
)

func GetBurper(directory string) (*burp.Application, *burper.Flow) {
	if !utils.HasSuffixStr(directory, ".toml") {
		directory = filepath.Join(directory, "burp.toml")
	}
	flow, err := burper.FromFile(directory)
	if err != nil {
		fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We couldn't analyze into ", console.Highlight, directory, " file!")
		fmt.Println(chalk.Red, err.Error())
		return nil, nil
	}
	var application burp.Application
	if err = toml.Unmarshal(flow.Bytes(), &application); err != nil {
		fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We couldn't analyze into ", console.Highlight, directory, " file!")
		fmt.Println(chalk.Red, err.Error())
		return nil, nil
	}
	return &application, flow
}

func GetEnvironmentFile(application *burp.Application) (*[]byte, bool) {
	if application.Environment.ServerSide {
		fmt.Println(chalk.Red, "⚠", chalk.Reset, "Server-side translation is enabled, skipping local translations of environment file.")
		return nil, true
	}
	translation, err := application.Environment.Translate("")
	if err != nil {
		fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We couldn't read the environment file!")
		fmt.Println(chalk.Red, err.Error())
		return nil, false
	}
	if application.Environment.Override {
		fmt.Println(chalk.Red, "⚠", chalk.Reset, "Environment file override is enabled, overriding ", console.Highlight,
			application.Environment.Baseline, chalk.Reset)
		if err = fileutils.Save(application.Environment.Baseline, []byte(*translation)); err != nil {
			fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We couldn't override the environment file!")
			fmt.Println(chalk.Red, err.Error())
			return nil, false
		}
	}
	return utils.Ptr([]byte(*translation)), true
}
