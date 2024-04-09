package middlewares

import (
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigFastest

func UseJsoniter(app *gin.Engine) {
	app.Use(func(c *gin.Context) {
		c.Set("json", json)
		c.Next()
	})
}
