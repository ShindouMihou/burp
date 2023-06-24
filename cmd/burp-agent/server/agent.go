package server

import (
	"burp/cmd/burp-agent/server/middlewares"
	"burp/cmd/burp-agent/server/responses"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"strconv"
)

// Executioner are tasks that are executed during Init which can modify the state of the Gin engine.
// You should use this when you want to add a new route, a group of routes, middlewares and other
// related that modifies the Gin state.
type Executioner = func(app *gin.Engine)

var Executioners []Executioner

// Init initializes all that is needed to start the burp-agent application, this involves checking the authentication
// properties, checking the TLS (SSL) certificates and running the Executioners which adds the routes, and other
// middlewares.
func Init(port int16) {
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
		if err := app.RunTLS(":"+strconv.FormatInt(int64(port), 10), cert, key); err != nil {
			log.Panic().Err(err).Msg("Cannot Start Gin")
			return
		}
	}()
	log.Info().
		Int16("port", port).
		Str("host", "0.0.0.0").
		Msg("Gin is now running")
}

// Add is a utility method that adds an Executioner to the Executioners stack, this returns a boolean
// that is always true, the return value is intended to allow us to call it using an empty variable name
// without much overhead.
func Add(executioner Executioner) bool {
	Executioners = append(Executioners, executioner)
	return true
}
