package handlers

import (
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
	"wios_server/conf"
	"wios_server/entity"
	"wios_server/handlers/msg"
	"wios_server/middlewares"
	"wios_server/service"
)

type RoleController struct {
	roleService *service.RoleService
}

func NewRoleController(app *gin.Engine) *RoleController {
	ctrl := &RoleController{roleService: service.NewRoleService(conf.Db)}
	group := app.Group("/role")
	hds := []middlewares.Auth{
		{Method: "POST", Group: "/role", Path: "/byPage", Auth: "role.list", Handler: ctrl.byPage},
		{Method: "POST", Group: "/role", Path: "/saveUpdate", Auth: "role.update", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/role", Path: "/delete", Auth: "role.delete", Handler: ctrl.delete},
	}
	for i, hd := range hds {
		middlewares.AuthMap[hd.Group+hd.Path] = &hds[i]
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}
func (ctrl *RoleController) byPage(ctx *gin.Context) {
	middlewares.ByPage(ctx, func(pageNumber int, pageSize int) (int64, any, error) {
		return ctrl.roleService.ListByPage(pageNumber, pageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.Role{}, "r")
		})
	})
}

func (ctrl *RoleController) saveUpdate(ctx *gin.Context) {
	var role entity.Role
	if err := ctx.ShouldBindJSON(&role); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.roleService.SaveUpdate(&role)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &role)
		return
	}
	msg.Error(ctx, "saveUpdate failed")
}

func (ctrl *RoleController) delete(ctx *gin.Context) {
	var role entity.Role
	if err := ctx.ShouldBindJSON(&role); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.roleService.Delete(&role)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &role)
		return
	}
	msg.Error(ctx, "delete failed")
}
