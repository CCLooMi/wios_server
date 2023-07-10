package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"wios_server/service"

	"github.com/gin-gonic/gin"
)

// 用户列表查询路由
func GetUserListHandler(db *sql.DB) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		// 获取请求参数
		pageNumber, _ := strconv.Atoi(ctx.Query("pageNumber"))
		pageSize, _ := strconv.Atoi(ctx.Query("pageSize"))

		// 查询用户列表
		userService := service.NewUserService(db)
		count, users, err := userService.FindByPage(pageNumber, pageSize)
		if err != nil {
			panic(err)
		}
		result := map[string]interface{}{
			"count": count,
			"users": users,
		}
		// 返回用户列表
		ctx.JSON(http.StatusOK, result)
	}
}
