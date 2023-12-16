package msg

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Ok(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, []any{0, data})
}
func Oks(ctx *gin.Context, data ...any) {
	ctx.JSON(http.StatusOK, append([]any{0}, data...))
}
func Error(ctx *gin.Context, msg any) {
	ctx.JSON(http.StatusOK, []any{1, msg})
}
