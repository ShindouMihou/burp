package routes

import (
	"burp/cmd/burp-agent/server"
	"burp/cmd/burp-agent/server/requests"
	responses2 "burp/cmd/burp-agent/server/responses"
	"burp/internal/docker"
	"burp/internal/services"
	"burp/pkg/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml/v2"
)

var _ = server.Add(func(app *gin.Engine) {
	app.POST("/application/stop", func(ctx *gin.Context) {
		logger := responses2.Logger(ctx)
		bytes, ok := requests.GetBurpFile(ctx)
		if !ok {
			return
		}
		logger.Info().Msg("Spawning a server-side stream...")
		responses2.AddSseHeaders(ctx)

		channel := utils.Ptr(make(chan any, 10))
		go func() {
			// IMPT: All deployments should be synchronous to  prevent an existential crisis
			// that doesn't exist, but still to be safe.
			responses2.ChannelSend(channel, responses2.CreateChannelOk("Waiting for deployment agent..."))
			logger.Info().Msg("Waiting for deployment agent...")

			lock.Lock()
			defer lock.Unlock()

			var burp services.Burp
			if err := toml.Unmarshal(bytes, &burp); err != nil {
				logger.Info().Err(err).Msg("Failed to parse TOML file into Burp services")
				responses2.ChannelSend(channel, responses2.CreateChannelError("Failed to parse TOML file into Burp services", err.Error()))
				return
			}

			responses2.ChannelSend(channel, responses2.CreateChannelOk("Killing main service container (burp."+burp.Service.Name+")...."))
			if err := docker.Kill(fmt.Sprint("burp.", burp.Service.Name)); err != nil {
				logger.Info().Err(err).Str("name", burp.Service.Name).Msg("Failed to kill main service container")
				responses2.ChannelSend(channel, responses2.CreateChannelError("Failed to kill main service container (burp."+burp.Service.Name+")", err.Error()))
				return
			}
			responses2.ChannelSend(channel, responses2.CreateChannelOk("Killed main service container (burp."+burp.Service.Name+")"))
			for _, dependency := range burp.Dependencies {
				responses2.ChannelSend(channel, responses2.CreateChannelOk("Killing dependency container (burp."+dependency.Name+")...."))
				if err := docker.Kill(fmt.Sprint("burp.", dependency.Name)); err != nil {
					logger.Info().Err(err).Str("name", dependency.Name).Msg("Failed to kill dependency container")
					responses2.ChannelSend(channel, responses2.CreateChannelError("Failed to kill dependency container (burp."+dependency.Name+")", err.Error()))
					return
				}
				responses2.ChannelSend(channel, responses2.CreateChannelOk("Killed dependency container (burp."+dependency.Name+")"))
			}
			defer close(*channel)
		}()

		responses2.Stream(ctx, channel)
	})
})
