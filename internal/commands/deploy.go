package commands

import (
	"burp/internal/api"
	"burp/internal/burpy"
	"burp/internal/commands/logins"
	"burp/pkg/console"
	"burp/pkg/fileutils"
	"bytes"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli/v2"
	"path/filepath"
)

var Deploy = &cli.Command{
	Name:        "deploy",
	Description: "Deploys an application to a remote server.",
	Action: func(ctx *cli.Context) error {
		var answers ServerRequestSurvey
		if err := survey.Ask(ServerRequestQuestions, &answers); err != nil {
			return err
		}
		console.Clear()
		answers.Keys.Sanitize()
		secrets, ok := logins.MustUnlock(&answers.Keys)
		if !ok {
			return nil
		}
		burp, tree := api.GetBurper(answers.Directory)
		if burp == nil || tree == nil {
			return nil
		}
		request := secrets.Client().
			EnableTrace().
			SetMultipartField("package[]", "burp.toml", "application/toml", bytes.NewBuffer(tree.Bytes())).
			SetDoNotParseResponse(true)

		if len(burp.Includes) > 0 {
			err := burpy.Package(burp)
			if err != nil {
				fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "Mom stopped us from packing our things to escape!")
				fmt.Println(chalk.Red, err.Error())
				return nil
			}

			fileName := fmt.Sprint(burp.Service.Name, "_includes.tar.gz")
			tarName := filepath.Join(burpy.TemporaryFilesFolder, ".packaged", fileName)

			pkg, err := fileutils.Open(tarName)
			if err != nil {
				fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "Mom stopped us from packing our things to escape!")
				fmt.Println(chalk.Red, err.Error())
				return nil
			}

			request = request.SetMultipartField("package[]", fileName, "application/gzip", pkg)
		}

		api.Streamed(request.Put(secrets.Link("application")))
		return nil
	},
}
