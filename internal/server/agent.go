package server

import (
	"burp/internal/server/middlewares"
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
		app.Use(logger.SetLogger(), middlewares.Authenticated)
		for _, executioner := range Executioners {
			executioner(app)
		}
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
