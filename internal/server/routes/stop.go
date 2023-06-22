package routes

import (
	"burp/internal/burper"
	"burp/internal/docker"
	"burp/internal/server"
	"burp/internal/server/mimes"
	"burp/internal/server/responses"
	"burp/internal/services"
	"burp/pkg/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml/v2"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

var _ = server.Add(func(app *gin.Engine) {
	app.DELETE("/application", func(ctx *gin.Context) {
		logger := responses.Logger(ctx)
		if ctx.ContentType() != "multipart/form-data" {
			responses.InvalidPayload.Reply(ctx)
			return
		}
		file, err := ctx.FormFile("burp")
		if err != nil {
			if errors.Is(err, http.ErrMissingFile) {
				responses.InvalidPayload.Reply(ctx)
				return
			}
			responses.HandleErr(ctx, err)
			return
		}
		contentType := file.Header.Get("Content-Type")
		if contentType != mimes.TOML_MIMETYPE {
			logger.Error().Str("Content-Type", contentType).Msg("Invalid Payload")
			responses.InvalidPayload.Reply(ctx)
			return
		}
		f, err := file.Open()
		if err != nil {
			responses.HandleErr(ctx, err)
			return
		}
		bytes, err := io.ReadAll(f)
		if err != nil {
			responses.HandleErr(ctx, err)
			return
		}

		logger.Info().Msg("Starting server-side stream...")
		ctx.Writer.Header().Set("Content-Type", "text/event-stream")
		ctx.Writer.Header().Set("Cache-Control", "no-cache")
		ctx.Writer.Header().Set("Connection", "keep-alive")
		ctx.Writer.Header().Set("Transfer-Encoding", "chunked")

		channel := utils.Ptr(make(chan any, 10))

		lock.TryLock()
		go func() {
			defer lock.Unlock()
			tree, err := burper.FromBytes(bytes)
			if err != nil {
				logger.Info().Err(err).Msg("Failed to parse TOML file into Burp tree")
				responses.ChannelSend(channel, responses.CreateChannelError("Failed to parse TOML file into Burp tree", err.Error()))
				return
			}

			var burp services.Burp
			if err = toml.Unmarshal(tree.Bytes(), &burp); err != nil {
				logger.Info().Err(err).Msg("Failed to parse TOML file into Burp services")
				responses.ChannelSend(channel, responses.CreateChannelError("Failed to parse TOML file into Burp services", err.Error()))
				return
			}

			responses.ChannelSend(channel, responses.CreateChannelOk("Killing main service container (burp."+burp.Service.Name+")...."))
			if err = docker.Kill(fmt.Sprint("burp.", burp.Service.Name)); err != nil {
				logger.Info().Err(err).Str("name", burp.Service.Name).Msg("Failed to kill main service container")
				responses.ChannelSend(channel, responses.CreateChannelError("Failed to kill main service container (burp."+burp.Service.Name+")", err.Error()))
				return
			}
			responses.ChannelSend(channel, responses.CreateChannelOk("Killed main service container (burp."+burp.Service.Name+")"))
			for _, dependency := range burp.Dependencies {
				responses.ChannelSend(channel, responses.CreateChannelOk("Killing dependency container (burp."+dependency.Name+")...."))
				if err = docker.Kill(fmt.Sprint("burp.", dependency.Name)); err != nil {
					logger.Info().Err(err).Str("name", dependency.Name).Msg("Failed to kill dependency container")
					responses.ChannelSend(channel, responses.CreateChannelError("Failed to kill dependency container (burp."+dependency.Name+")", err.Error()))
					return
				}
				responses.ChannelSend(channel, responses.CreateChannelOk("Killed dependency container (burp."+dependency.Name+")"))
			}
			defer close(*channel)
		}()

		ctx.Stream(func(w io.Writer) bool {
			if msg, ok := <-*channel; ok {
				log.Info().Any("data", msg).Msg("Received stream message")
				ctx.SSEvent("data", msg)
				return true
			}
			return false
		})
	})
})
