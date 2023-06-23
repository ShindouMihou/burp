package server

import (
	"burp/cmd/burp-agent/server/middlewares"
	"burp/cmd/burp-agent/server/responses"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Executioner = func(app *gin.Engine)

var Executioners []Executioner

func Init() {
	EnsureAuthentication()
	cert, key, err := GetSsl()
	if err != nil {
		log.Panic().Err(err).Msg("Failed to generate, or retrieve SSL certificates.")
		return
	}
	gin.SetMode(gin.ReleaseMode)
	go func() {
		app := gin.New()
		// IMPRT: Limit file uploads to 5 MiB since the file upload mechanism
		// of Burp is intended for additional configuration files, not massive
		// dangerous binaries.
		app.MaxMultipartMemory = 5 << 20
		app.Use(logger.SetLogger(), middlewares.Authenticated)
		for _, executioner := range Executioners {
			executioner(app)
		}
		app.NoRoute(func(ctx *gin.Context) { responses.NotFound.Reply(ctx) })
		if err := app.RunTLS(":8873", cert, key); err != nil {
			log.Panic().Err(err).Msg("Cannot Start Gin")
			return
		}
	}()
	log.Info().
		Int16("port", 8873).
		Str("host", "0.0.0.0").
		Msg("Gin is now running")
}

func Add(executioner Executioner) bool {
	Executioners = append(Executioners, executioner)
	return true
}
