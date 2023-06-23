package routes

import (
	"burp/internal/docker"
	"burp/internal/server"
	"burp/internal/server/requests"
	"burp/internal/server/responses"
	"burp/internal/services"
	"burp/pkg/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml/v2"
)

var _ = server.Add(func(app *gin.Engine) {
	app.POST("/application/start", func(ctx *gin.Context) {
		logger := responses.Logger(ctx)
		bytes, ok := requests.GetBurpFile(ctx)
		if !ok {
			return
		}
		logger.Info().Msg("Starting server-side stream...")
		responses.AddSseHeaders(ctx)

		channel := utils.Ptr(make(chan any, 10))
		go func() {
			// IMPT: All deployments should be synchronous to  prevent an existential crisis
			// that doesn't exist, but still to be safe.
			responses.ChannelSend(channel, responses.CreateChannelOk("Waiting for deployment agent..."))
			logger.Info().Msg("Waiting for deployment agent...")

			lock.Lock()
			defer lock.Unlock()
			var burp services.Burp
			if err := toml.Unmarshal(bytes, &burp); err != nil {
				logger.Info().Err(err).Msg("Failed to parse TOML file into Burp services")
				responses.ChannelSend(channel, responses.CreateChannelError("Failed to parse TOML file into Burp services", err.Error()))
				return
			}

			for _, dependency := range burp.Dependencies {
				responses.ChannelSend(channel, responses.CreateChannelOk("Starting dependency container (burp."+dependency.Name+")...."))
				if err := docker.Start(fmt.Sprint("burp.", dependency.Name)); err != nil {
					logger.Info().Err(err).Str("name", dependency.Name).Msg("Failed to start dependency container")
					responses.ChannelSend(channel, responses.CreateChannelError("Failed to start dependency container (burp."+dependency.Name+")", err.Error()))
					return
				}
				responses.ChannelSend(channel, responses.CreateChannelOk("Started dependency container (burp."+dependency.Name+")"))
			}

			responses.ChannelSend(channel, responses.CreateChannelOk("Starting main service container (burp."+burp.Service.Name+")...."))
			if err := docker.Start(fmt.Sprint("burp.", burp.Service.Name)); err != nil {
				logger.Info().Err(err).Str("name", burp.Service.Name).Msg("Failed to start main service container")
				responses.ChannelSend(channel, responses.CreateChannelError("Failed to start main service container (burp."+burp.Service.Name+")", err.Error()))
				return
			}
			responses.ChannelSend(channel, responses.CreateChannelOk("Started main service container (burp."+burp.Service.Name+")"))
			defer close(*channel)
		}()

		responses.Stream(ctx, channel)
	})
})
