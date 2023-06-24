package commands

import (
	"burp/cmd/burp-agent/server/mimes"
	api "burp/cmd/burp/api"
	"burp/cmd/burp/commands/logins"
	"burp/cmd/burp/commands/templates"
	"burp/internal/burpy"
	"burp/internal/services"
	"burp/pkg/console"
	"burp/pkg/fileutils"
	"bytes"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/go-resty/resty/v2"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli/v2"
	"path/filepath"
)

var Deploy = &cli.Command{
	Name:        "deploy",
	Description: "Deploys an application to a remote server.",
	Action: func(ctx *cli.Context) error {
		var answers templates.ServerRequestSurvey
		if err := survey.Ask(templates.ServerRequestQuestions, &answers); err != nil {
			return err
		}
		console.Clear()
		answers.Keys.Sanitize()
		secrets, ok := logins.MustUnlock(&answers.Keys)
		if !ok {
			return nil
		}
		burp, flow := api.GetBurper(answers.Directory)
		if burp == nil || flow == nil {
			return nil
		}
		environmentFile, ok := api.GetEnvironmentFile(burp)
		if !ok {
			return nil
		}
		request := secrets.Client().
			EnableTrace().
			SetMultipartField("package[]", "burp.toml", "application/toml", bytes.NewReader(flow.Bytes())).
			SetDoNotParseResponse(true)
		if ok && environmentFile != nil {
			request = request.SetMultipartField("package[]", ".env", mimes.TEXT_MIMETYPE, bytes.NewReader(*environmentFile))
		}
		if ok := Package(burp, request); !ok {
			return nil
		}
		api.Streamed(request.Put(secrets.Link("application")))
		return nil
	},
}

func Package(burp *services.Burp, request *resty.Request) bool {
	if len(burp.Includes) > 0 {
		err := burpy.Package(burp)
		if err != nil {
			fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "Mom stopped us from packing our things to escape!")
			fmt.Println(chalk.Red, err.Error())
			return false
		}

		fileName := fmt.Sprint(burp.Service.Name, "_includes.tar.gz")
		tarName := filepath.Join(burpy.TemporaryFilesFolder, ".packaged", fileName)

		pkg, err := fileutils.Open(tarName)
		if err != nil {
			fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "Mom stopped us from packing our things to escape!")
			fmt.Println(chalk.Red, err.Error())
			return false
		}

		request = request.SetMultipartField("package[]", fileName, "application/gzip", pkg)
	}
	return true
}
