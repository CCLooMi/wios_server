package handlers

import (
	"wios_server/conf"
	"wios_server/entity"
	"wios_server/handlers/msg"
	"wios_server/middlewares"
	"wios_server/service"

	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
)

type MenuController struct {
	menuService *service.MenuService
}

func NewMenuController(app *gin.Engine) *MenuController {
	ctrl := &MenuController{menuService: service.NewMenuService(conf.Db)}
	group := app.Group("/menu")
	hds := []middlewares.Auth{
		{Method: "POST", Group: "/menu", Path: "/byPage", Auth: "menu.list", Handler: ctrl.byPage},
		{Method: "POST", Group: "/menu", Path: "/saveUpdate", Auth: "menu.update", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/menu", Path: "/delete", Auth: "menu.delete", Handler: ctrl.delete},
	}
	for _, hd := range hds {
		middlewares.AuthMap[hd.Group+hd.Path] = &hd
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}

func (ctrl *MenuController) byPage(ctx *gin.Context) {
	middlewares.ByPage(ctx, func(pageNumber int, pageSize int) (int64, any, error) {
		return ctrl.menuService.ListByPage(pageNumber, pageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.Menu{}, "m")
		})
	})
}

func (ctrl *MenuController) saveUpdate(ctx *gin.Context) {
	var menu entity.Menu
	if err := ctx.ShouldBindJSON(&menu); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.menuService.SaveUpdate(&menu)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &menu)
		return
	}
	msg.Error(ctx, "save failed")
}

func (ctrl *MenuController) delete(ctx *gin.Context) {
	var menu entity.Menu
	if err := ctx.ShouldBindJSON(&menu); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.menuService.Delete(&menu)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &menu)
		return
	}
	msg.Error(ctx, "delete failed")
}
