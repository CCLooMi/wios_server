package handlers

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
	"strconv"
	"wios_server/entity"
	"wios_server/handlers/msg"
	"wios_server/service"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(app *gin.Engine, db *sql.DB) *UserController {
	ctrl := &UserController{userService: service.NewUserService(db)}
	group := app.Group("/user")
	group.POST("/listByPage", ctrl.byPage)
	return ctrl
}

func (ctrl *UserController) byPage(ctx *gin.Context) {
	// 获取请求参数
	pageNumber, _ := strconv.Atoi(ctx.Query("pageNumber"))
	pageSize, _ := strconv.Atoi(ctx.Query("pageSize"))
	count, users, err := ctrl.userService.ListByPage(pageNumber, pageSize, func(sm *mak.SQLSM) {
		sm.SELECT("*").FROM(entity.User{}, "u")
	})
	if err != nil {
		panic(err)
	}
	result := map[string]interface{}{
		"count": count,
		"data":  users,
	}
	msg.Ok(ctx, result)
}
