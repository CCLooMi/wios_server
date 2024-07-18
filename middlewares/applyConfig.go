package middlewares

import (
	"github.com/gin-gonic/gin"
	"wios_server/conf"
)

func ApplyConfig(c *gin.Context, config *conf.Config) {
	for key, value := range config.Header {
		c.Writer.Header().Set(key, value)
	}
	if config.EnableCORS {
		c.Writer.Header().Set("Access-Control-Allow-Methods", c.Request.Method)
		c.Writer.Header().Set("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(200)
		return
	}
	c.Next()
}
