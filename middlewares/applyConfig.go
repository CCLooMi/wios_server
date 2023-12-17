package middlewares

import (
	"github.com/gin-gonic/gin"
	"wios_server/conf"
)

func ApplyConfig(c *gin.Context) {
	for key, value := range conf.Cfg.Header {
		c.Writer.Header().Set(key, value)
	}
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(200)
		return
	}
	c.Next()
}
