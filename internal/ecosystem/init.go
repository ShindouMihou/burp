package ecosystem

import (
	"burp/internal/auth"
	"burp/internal/burper/functions"
	"burp/internal/server/routes"
	"burp/pkg/env"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

func Init() {
	_ = godotenv.Load()

	functions.RegisterFunctions()
	routes.Discover()

	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Stack()
	if strings.EqualFold(env.AgentMode.Or("release"), "debug") {
		logger = logger.Caller()
	}
	log.Logger = logger.Logger()

	auth.Load()
}
