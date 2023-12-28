package middlewares

import (
	"github.com/gin-gonic/gin"
	"wios_server/conf"
)

func ApplyConfig(c *gin.Context) {
	for key, value := range conf.Cfg.Header {
		c.Writer.Header().Set(key, value)
	}
	if conf.Cfg.EnableCORS {
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
