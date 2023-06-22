package middlewares

import (
	"burp/internal/server/responses"
	"burp/pkg/env"
	"burp/pkg/utils"
	"github.com/alexedwards/argon2id"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"strings"
)

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