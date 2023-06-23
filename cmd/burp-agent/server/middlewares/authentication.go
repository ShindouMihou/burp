package middlewares

import (
	responses "burp/cmd/burp-agent/server/responses"
	"burp/pkg/env"
	"burp/pkg/utils"
	"github.com/alexedwards/argon2id"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"strings"
)

// Authenticated is a middleware that checks the authentication of a request.
// It checks for the X-Burp-Signature first before proceeding with the Authorization
// otherwise known as the Burp Secret.
//
// If it fails in either of the checks, then the middleware will stop them with a 401
// HTTP status code.
var Authenticated gin.HandlerFunc = func(ctx *gin.Context) {
	signature := ctx.GetHeader("X-Burp-Signature")
	if signature == "" || signature != env.BurpSignature.Get() {
		responses.Unauthorized.Reply(ctx)
		return
	}
	authorization := ctx.GetHeader("Authorization")
	if authorization == "" || !utils.HasPrefixStr(authorization, "Bearer ") {
		responses.Unauthorized.Reply(ctx)
		return
	}
	token := strings.SplitN(authorization, " ", 2)[1]
	log.Debug().Str("token", token).Msg("Received Authorization")
	ok, err := argon2id.ComparePasswordAndHash(token, env.BurpSecret.Get())
	if err != nil {
		responses.HandleErr(ctx, err)
		return
	}
	if !ok {
		responses.Unauthorized.Reply(ctx)
		return
	}
	ctx.Next()
}
