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

type PermissionController struct {
	permissionService *service.PermissionService
}

func NewPermissionController(app *gin.Engine) *PermissionController {
	ctrl := &PermissionController{permissionService: service.NewPermissionService(conf.Db)}
	group := app.Group("/permission")
	hds := []middlewares.Auth{
		{Method: "POST", Group: "/permission", Path: "/byPage", Auth: "permission.list", Handler: ctrl.byPage},
		{Method: "POST", Group: "/permission", Path: "/saveUpdate", Auth: "permission.update", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/permission", Path: "/delete", Auth: "permission.delete", Handler: ctrl.delete},
	}
	for i, hd := range hds {
		middlewares.AuthMap[hd.Group+hd.Path] = &hds[i]
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}

func (ctrl *PermissionController) byPage(ctx *gin.Context) {
	middlewares.ByPage(ctx, func(pageNumber int, pageSize int) (int64, any, error) {
		return ctrl.permissionService.ListByPage(pageNumber, pageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.Permission{}, "p")
		})
	})
}

func (ctrl *PermissionController) saveUpdate(ctx *gin.Context) {
	var permission entity.Permission
	if err := ctx.ShouldBindJSON(&permission); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.permissionService.SaveUpdate(&permission)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &permission)
		return
	}
	msg.Error(ctx, "saveUpdate failed")
}

func (ctrl *PermissionController) delete(ctx *gin.Context) {
	var permission entity.Permission
	if err := ctx.ShouldBindJSON(&permission); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.permissionService.Delete(&permission)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &permission)
		return
	}
	msg.Error(ctx, "delete failed")
}
