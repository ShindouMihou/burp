package commands

import (
	"burp/internal/burper"
	"burp/pkg/console"
	"fmt"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli/v2"
)

var Eval = &cli.Command{
	Name:        "eval",
	Description: "Evaluates the given argument with the Burper engine.",
	Action: func(ctx *cli.Context) error {
		arg := ctx.Args().First()
		if arg == "" {
			fmt.Println(chalk.Red, "૮₍˃⤙˂₎ა", chalk.Reset, "You need to supply the ", console.Highlight, "text", chalk.Reset, "to evaluate!")
			return nil
		}
		flow, err := burper.FromString(arg)
		if err != nil {
			fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "Failed to decode the secret nuclear codes!")
			fmt.Println(chalk.Red, err.Error())
			return nil
		}
		fmt.Println(flow.String())
		return nil
	},
}
