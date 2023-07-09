package templates

import (
	"burp/cmd/burp/api"
	"burp/cmd/burp/commands/logins"
	"burp/pkg/console"
	"burp/pkg/fileutils"
	"burp/pkg/utils"
	"bytes"
	"errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/go-resty/resty/v2"
	"github.com/urfave/cli/v2"
	"path/filepath"
)

var ServerRequestQuestions = []*survey.Question{
	{
		Name: "name",
		Prompt: &survey.Select{
			Message: "On which server do you want to perform this action? (case-insensitive)",
			Help:    "If you haven't created one, use the `burp login` command.",
			Options: []string{},
		},
		Validate: survey.ComposeValidators(
			survey.Required,
			func(ans interface{}) error {
				file := fileutils.Sanitize(ans.(core.OptionAnswer).Value)
				file = filepath.Join(logins.Folder, file+".json")

				exists, err := utils.Exists(file)
				if err != nil {
					return err
				}
				if !exists {
					return errors.New("cannot find credentials for that server, try adding it back again")
				}
				return nil
			}),
	},
	{
		Name: "encryption",
		Prompt: &survey.Password{
			Message: "Enter the encryption key for this server. (16 characters minimum)",
			Help:    "This will be used to decrypt the credentials of the server.",
		},
		Validate: survey.MinLength(16),
	},
	{
		Name: "directory",
		Prompt: &survey.Input{
			Message: "Which directory from here has the Burp.toml? (relative, defaults to current)",
			Help:    "For more information, please refer to https://github.com/ShindouMihou/burp.",
		},
		Validate: func(ans interface{}) error {
			file := ans.(string)
			wasDirectory := false
			if !utils.HasSuffixStr(file, ".toml") {
				file = filepath.Join(file, "burp.toml")
				wasDirectory = true
			}

			exists, err := utils.Exists(file)
			if err != nil {
				return err
			}
			if !exists {
				if wasDirectory {
					return errors.New("there is no burp.toml found in that directory")
				}
				return errors.New("the file specified cannot be found")
			}
			return nil
		},
	},
}

type ServerRequestSurvey struct {
	api.Keys
	Directory string `json:"directory"`
}

type ServerRequestAction = func(secrets *api.Secrets, request *resty.Request) (*resty.Response, error)

// CreateServerRequestCommand creates a command that involves decrypting the secret credentials
// and performing a request to the server, but this is limited to only commands that require uploading
// a Burp.toml and only a Burp.toml (no additional files).
func CreateServerRequestCommand(name string, description string, action ServerRequestAction) *cli.Command {
	return &cli.Command{
		Name:        name,
		Description: description,
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
			burp, flow := api.GetBurper(answers.Directory)
			if burp == nil || flow == nil {
				return nil
			}
			client, ok := secrets.ClientWithTls(answers.Keys.Name)
			if !ok {
				return nil
			}
			request := client.EnableTrace().
				SetMultipartField("burp", "burp.toml", "application/toml", bytes.NewReader(flow.Bytes())).
				SetDoNotParseResponse(true)

			api.Streamed(action(secrets, request))
			return nil
		},
	}
}
