package commands

import "github.com/urfave/cli/v2"

var App = &cli.App{
	Name:        "burp",
	Description: "Deploying smaller applications shouldn't be crazy complicated.",
	Commands:    []*cli.Command{Login, Logout, Hash, Deploy, Start, Stop, Remove, Restart, Here},
	Authors: []*cli.Author{
		{
			Name:  "Shindou Mihou",
			Email: "hello@mihou.pw",
		},
	},
}
