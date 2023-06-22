package routes

import (
	"burp/internal/server"
	"github.com/gin-gonic/gin"
	"net/http"
)

var _ = server.Add(func(app *gin.Engine) {
	app.GET("/", func(context *gin.Context) {
		context.Status(http.StatusNoContent)
	})
})
