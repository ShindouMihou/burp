package app

import (
	"burp/internal/auth"
	"burp/internal/burper/functions"
	"burp/internal/server/routes"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func Init() {
	functions.RegisterFunctions()
	routes.Discover()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Msg("Loading environment configuration")
	_ = godotenv.Load()
	auth.Load()
}
