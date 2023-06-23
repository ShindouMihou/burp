package commands

import (
	"burp/pkg/console"
	"fmt"
	"github.com/alexedwards/argon2id"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli/v2"
)

var Hash = &cli.Command{
	Name:        "hash",
	Description: "Hashes a given text with argon2id that can be used with Burp.",
	Action: func(ctx *cli.Context) error {
		arg := ctx.Args().First()
		if arg == "" {
			fmt.Println(chalk.Red, "૮₍˃⤙˂₎ა", chalk.Reset, "You need to supply the ", console.Highlight, "text", chalk.Reset, "to hash!")
			return nil
		}
		hash, err := argon2id.CreateHash(arg, argon2id.DefaultParams)
		if err != nil {
			fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "We failed to negotiate with the operating system!")
			fmt.Println(chalk.Red, err.Error())
			return nil
		}
		fmt.Println(hash)
		return nil
	},
}
