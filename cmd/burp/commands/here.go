package commands

import (
	"burp/cmd/burp-agent/server"
	"burp/cmd/burp-agent/server/mimes"
	"burp/cmd/burp/api"
	"burp/internal/docker"
	"burp/pkg/console"
	"burp/pkg/env"
	"burp/pkg/shutdown"
	"bytes"
	"fmt"
	"github.com/alexedwards/argon2id"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var Here = &cli.Command{
	Name: "here",
	Description: "Deploys an application locally, this is recommended only for installing the Burp agent. It opens a local " +
		"Burp agent's server to handle the transactions.",
	Usage: "burp deploy [directory, defaults to working directory]",
	Action: func(ctx *cli.Context) error {
		directory := ctx.Args().First()
		burp, flow := api.GetBurper(directory)
		if burp == nil || flow == nil {
			return nil
		}
		environmentFile, ok := api.GetEnvironmentFile(burp)
		if !ok {
			return nil
		}
		if err := docker.Init(); err != nil {
			fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "Failed to connect to Docker!")
			fmt.Println(chalk.Red, err.Error())
			return nil
		}
		hash, err := argon2id.CreateHash("local", argon2id.DefaultParams)
		if err != nil {
			fmt.Println(chalk.Red, "(◞‸◟；)", chalk.Reset, "Failed to create temporary secret token!")
			fmt.Println(chalk.Red, err.Error())
			return nil
		}
		_ = os.Setenv(env.BurpSecret.String(), hash)
		_ = os.Setenv(env.BurpSignature.String(), "local")
		_ = os.Setenv(env.AgentMode.String(), "local")

		secrets := &api.Secrets{
			Server:    "https://localhost:8875",
			Secret:    "local",
			Signature: "local",
		}

		end := make(chan bool)
		go func() {
			log.Logger = log.Logger.Level(zerolog.ErrorLevel)
			fmt.Println(chalk.Red, "⚠", chalk.Reset, "You are about to deploy the application ", console.Highlight, burp.Service.Name,
				chalk.Reset, " locally. You have 5 seconds to cancel (CTRL+C) the deployment.")
			server.Init(8875)
			fmt.Println(chalk.Green, "✓", chalk.Reset, "Burp's Agent server is now temporarily running locally to accommodate this deployment.")
		}()
		go func() {
			time.Sleep(5 * time.Second)
			request := api.CreateInsecureWith(secrets).
				EnableTrace().
				SetMultipartField("package[]", "burp.toml", "application/toml", bytes.NewReader(flow.Bytes())).
				SetDoNotParseResponse(true)
			if ok && environmentFile != nil {
				request = request.SetMultipartField("package[]", ".env", mimes.TEXT_MIMETYPE, bytes.NewReader(*environmentFile))
			}
			if ok := Package(burp, request); !ok {
				return
			}
			api.Streamed(request.Put(secrets.Link("application")))
			end <- true
		}()
		go func() {
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			<-sigs

			signal.Stop(sigs)
			end <- true
		}()
		<-end
		_ = shutdown.Cleanup()
		return nil
	},
}
