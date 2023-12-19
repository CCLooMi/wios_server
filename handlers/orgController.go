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

type OrgController struct {
	orgService *service.OrgService
}

func NewOrgController(app *gin.Engine) *OrgController {
	ctrl := &OrgController{orgService: service.NewOrgService(conf.Db)}
	group := app.Group("/org")
	hds := []middlewares.Auth{
		{Method: "POST", Group: "/org", Path: "/byPage", Auth: "org.list", Handler: ctrl.byPage},
		{Method: "POST", Group: "/org", Path: "/saveUpdate", Auth: "org.update", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/org", Path: "/delete", Auth: "org.delete", Handler: ctrl.delete},
	}
	for i, hd := range hds {
		middlewares.AuthMap[hd.Group+hd.Path] = &hds[i]
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}

func (ctrl *OrgController) byPage(ctx *gin.Context) {
	middlewares.ByPage(ctx, func(pageNumber int, pageSize int) (int64, any, error) {
		return ctrl.orgService.ListByPage(pageNumber, pageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.Org{}, "o")
		})
	})
}

func (ctrl *OrgController) saveUpdate(ctx *gin.Context) {
	var org entity.Org
	if err := ctx.ShouldBindJSON(&org); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.orgService.SaveUpdate(&org)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &org)
		return
	}
	msg.Error(ctx, "saveUpdate failed")
}

func (ctrl *OrgController) delete(ctx *gin.Context) {
	var org entity.Org
	if err := ctx.ShouldBindJSON(&org); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.orgService.Delete(&org)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &org)
		return
	}
	msg.Error(ctx, "delete failed")
}
