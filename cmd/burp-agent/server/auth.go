package server

import (
	"burp/pkg/env"
	"github.com/alexedwards/argon2id"
	"github.com/dchest/uniuri"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

// EnsureAuthentication ensures that we have a BurpSignature and a BurpSecret to protect our application.
// If it cannot find one and if we are in debug mode, then it generates one and logs them into console, otherwise,
// it will panic.
func EnsureAuthentication() {
	if env.BurpSecret.OrNull() == nil || env.BurpSignature.OrNull() == nil {
		if strings.EqualFold(env.AgentMode.Or("release"), "debug") {
			log.Panic().
				Str("BURP_SECRET", "argon2id_hashed string").
				Str("BURP_SIGNATURE", "string").
				Msg("Missing configuration properties")
		}
		token := uniuri.NewLen(64)
		signature := uniuri.NewLen(64)
		log.Warn().
			Str("BURP_SECRET", token).
			Str("BURP_SIGNATURE", signature).
			Msg("Missing configuration properties, but since  Debug mode is enabled, generated credentials needed.")
		hash, err := argon2id.CreateHash(token, argon2id.DefaultParams)
		if err != nil {
			log.Panic().
				Err(err).
				Msg("Failed to hash secret token.")
		}
		_ = os.Setenv(env.BurpSecret.String(), hash)
		_ = os.Setenv(env.BurpSignature.String(), signature)

		log.Warn().
			Str("BURP_SECRET", hash).
			Msg("Generated Hash for BURP_SECRET")
	}
}
