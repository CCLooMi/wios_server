package middleware

import (
    "github.com/kataras/iris/v12"
)

func LoggerMiddleware(ctx iris.Context) {
    // 记录请求日志
    ctx.Application().Logger().Infof("%s %s", ctx.Method(), ctx.Path())

    // 继续处理请求
    ctx.Next()
}

