package middlewares

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"wios_server/handlers/msg"
)

func ByPage(ctx *gin.Context, f func(pageNumber int, pageSize int) (int64, any, error)) {
	pageNumber, _ := strconv.Atoi(ctx.Query("pageNumber"))
	pageSize, _ := strconv.Atoi(ctx.Query("pageSize"))
	count, data, err := f(pageNumber, pageSize)
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	result := map[string]interface{}{
		"count": count,
		"data":  data,
	}
	msg.Ok(ctx, result)
}
