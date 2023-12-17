package handlers

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
	"wios_server/entity"
	"wios_server/middlewares"
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
	middlewares.ByPage(ctx, func(pageNumber int, pageSize int) (int64, any, error) {
		return ctrl.userService.ListByPage(pageNumber, pageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.User{}, "u")
		})
	})
}
