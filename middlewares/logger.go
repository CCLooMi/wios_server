package middlewares

import (
	"log"

	"github.com/gin-gonic/gin"
)

func LoggerRequestInfo(ctx *gin.Context) {
	// 记录请求日志
	log.Printf("%s %s", ctx.Request.Method, ctx.Request.URL.Path)
	// 继续处理请求
	ctx.Next()
}
